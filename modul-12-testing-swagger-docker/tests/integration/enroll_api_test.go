//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/helper"
	"api-students/route"
)

type api struct {
	app *fiber.App
	jwt *helper.JWTManager
}

func newAPI(t *testing.T) api {
	t.Helper()
	raw, err := repository.NewRoleRepository(pool).LoadPermissions(context.Background())
	if err != nil {
		t.Fatalf("memuat permission: %v", err)
	}
	perms := helper.NewPermissionSet(raw)
	jwtManager := helper.NewJWTManager("secret-khusus-test-minimal-32-karakter", "test", time.Minute)
	courseService := service.NewCourseService(
		repository.NewCourseRepository(pool),
		repository.NewOutboxRepository(pool),
		repository.NewTxManager(pool),
	)
	app := config.NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)), route.Dependencies{
		Pool:          pool,
		JWT:           jwtManager,
		Permissions:   perms,
		CourseService: courseService,
	})
	return api{app: app, jwt: jwtManager}
}

func (a api) enroll(t *testing.T, student model.User, courseID int) (int, string) {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0, ""
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/"+strconv.Itoa(courseID)+"/enroll", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request enroll: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()
	var body struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Code
}

func (a api) drop(t *testing.T, student model.User, courseID int) int {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/courses/"+strconv.Itoa(courseID)+"/enroll", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request drop: %v", err)
		return 0
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func (a api) swap(t *testing.T, student model.User, fromID, toID int) (int, string) {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.AuthUser{UserID: student.ID, Username: student.Username, Role: "user"})
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0, ""
	}
	body, _ := json.Marshal(model.SwapRequest{FromCourseID: fromID, ToCourseID: toID})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments/swap", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request swap: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()
	var respBody struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	return resp.StatusCode, respBody.Code
}

func countRows(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return n
}

func TestEnrollEndpoint_WritesOutboxOnlyOnSuccess(t *testing.T) {
	a := newAPI(t)
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)
	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("pendaftaran pertama: status = %d, seharusnya 201", status)
	}
	if status, code := a.enroll(t, student, course.ID); status != http.StatusConflict || code != "ALREADY_ENROLLED" {
		t.Fatalf("pendaftaran kedua: dapat %d %s, seharusnya 409 ALREADY_ENROLLED", status, code)
	}
	events := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentCreated)
	if events != 1 {
		t.Errorf("event enrollment.created = %d, seharusnya 1", events)
	}
}

func TestEnrollEndpoint_ConcurrentCreditLimit(t *testing.T) {
	a := newAPI(t)
	student := seedStudent(t)
	const attempts = 10
	courses := make([]model.Course, attempts)
	for i := range courses {
		courses[i] = seedCourse(t, 3, 40)
	}
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = map[int]int{}
		start   = make(chan struct{})
	)
	for _, c := range courses {
		wg.Add(1)
		go func(courseID int) {
			defer wg.Done()
			<-start
			status, _ := a.enroll(t, student, courseID)
			mu.Lock()
			results[status]++
			mu.Unlock()
		}(c.ID)
	}
	close(start)
	wg.Wait()
	maxCourses := service.MaxCreditsPerSemester / 3
	if results[http.StatusCreated] != maxCourses {
		t.Errorf("berhasil = %d, seharusnya %d (hasil lengkap: %v)", results[http.StatusCreated], maxCourses, results)
	}
	enrolled := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE user_id = $1`, student.ID)
	if enrolled != maxCourses {
		t.Errorf("baris enrollments = %d, seharusnya %d", enrolled, maxCourses)
	}
}

func TestEnrollEndpoint_SwapTargetFull(t *testing.T) {
	a := newAPI(t)
	ctx := context.Background()
	student := seedStudent(t)
	fromCourse := seedCourse(t, 3, 40)
	toCourse := seedCourse(t, 3, 1)

	if status, _ := a.enroll(t, student, fromCourse.ID); status != http.StatusCreated {
		t.Fatalf("enroll kelas asal: status = %d, seharusnya 201", status)
	}

	otherStudent := seedStudent(t)
	if status, _ := a.enroll(t, otherStudent, toCourse.ID); status != http.StatusCreated {
		t.Fatalf("enroll kelas tujuan (penuh): status = %d, seharusnya 201", status)
	}

	coursesRepo := repository.NewCourseRepository(pool)

	status, code := a.swap(t, student, fromCourse.ID, toCourse.ID)
	if status != http.StatusConflict || code != "COURSE_FULL" {
		t.Errorf("swap ke kelas penuh: dapat status %d code %q, seharusnya 409 COURSE_FULL", status, code)
	}

	enrolledFrom := countRows(t, "SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2", student.ID, fromCourse.ID)
	if enrolledFrom != 1 {
		t.Errorf("mahasiswa harus tetap terdaftar di kelas asal: count = %d, seharusnya 1", enrolledFrom)
	}

	enrolledTo := countRows(t, "SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2", student.ID, toCourse.ID)
	if enrolledTo != 0 {
		t.Errorf("mahasiswa tidak boleh terdaftar di kelas tujuan: count = %d, seharusnya 0", enrolledTo)
	}

	fromDB, err := coursesRepo.FindByID(ctx, fromCourse.ID)
	if err != nil {
		t.Fatalf("FindByID kelas asal: %v", err)
	}
	if fromDB.EnrolledCount != 1 {
		t.Errorf("enrolled_count kelas asal = %d, seharusnya 1", fromDB.EnrolledCount)
	}

	toDB, err := coursesRepo.FindByID(ctx, toCourse.ID)
	if err != nil {
		t.Fatalf("FindByID kelas tujuan: %v", err)
	}
	if toDB.EnrolledCount != 1 {
		t.Errorf("enrolled_count kelas tujuan = %d, seharusnya 1", toDB.EnrolledCount)
	}

	events := countRows(t, "SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2", student.ID, model.EventEnrollmentSwapped)
	if events != 0 {
		t.Errorf("outbox tidak boleh ada event swapped: count = %d, seharusnya 0", events)
	}
}

func TestEnrollEndpoint_DropSuccess(t *testing.T) {
	a := newAPI(t)
	ctx := context.Background()
	student := seedStudent(t)
	course := seedCourse(t, 3, 40)

	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("enroll kelas: status = %d, seharusnya 201", status)
	}

	coursesRepo := repository.NewCourseRepository(pool)
	beforeDB, err := coursesRepo.FindByID(ctx, course.ID)
	if err != nil {
		t.Fatalf("FindByID sebelum drop: %v", err)
	}

	status := a.drop(t, student, course.ID)
	if status != http.StatusNoContent {
		t.Errorf("drop status = %d, seharusnya 204", status)
	}

	enrolledAfter := countRows(t, "SELECT COUNT(*) FROM enrollments WHERE user_id = $1 AND course_id = $2", student.ID, course.ID)
	if enrolledAfter != 0 {
		t.Errorf("baris enrollment = %d, seharusnya 0 setelah drop", enrolledAfter)
	}

	afterDB, err := coursesRepo.FindByID(ctx, course.ID)
	if err != nil {
		t.Fatalf("FindByID setelah drop: %v", err)
	}
	if afterDB.EnrolledCount != beforeDB.EnrolledCount-1 {
		t.Errorf("enrolled_count setelah drop = %d, seharusnya %d", afterDB.EnrolledCount, beforeDB.EnrolledCount-1)
	}

	droppedEvents := countRows(t, "SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2", student.ID, model.EventEnrollmentDropped)
	if droppedEvents != 1 {
		t.Errorf("event enrollment.dropped di outbox = %d, seharusnya 1", droppedEvents)
	}
}

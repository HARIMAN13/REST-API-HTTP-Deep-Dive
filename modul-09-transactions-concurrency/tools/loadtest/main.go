package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"api-students/app/model"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
)

func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}

	close(start)
	wg.Wait()
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

func main() {
	mode := flag.String("mode", "crowd", "Mode pengujian: crowd | credits | swap | supervisor | supervisor-switch")
	baseURL := flag.String("url", "http://localhost:3000", "Base URL server")
	flag.Parse()

	config.LoadEnv()
	jwtSecret := config.GetEnv("JWT_SECRET", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	jwtManager := helper.NewJWTManager(jwtSecret, "praktikum-backend", 15*time.Minute)

	pool, err := database.NewPool(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal terhubung ke database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctx := context.Background()

	type UserInfo struct {
		ID       int
		Username string
	}
	var testUsers []UserInfo
	rows, err := pool.Query(ctx, "SELECT id, username FROM users WHERE username LIKE 'mhs%' ORDER BY username")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gagal mengambil akun uji: %v\n", err)
		os.Exit(1)
	}
	for rows.Next() {
		var u UserInfo
		if err := rows.Scan(&u.ID, &u.Username); err == nil {
			testUsers = append(testUsers, u)
		}
	}
	rows.Close()

	if len(testUsers) == 0 {
		fmt.Fprintf(os.Stderr, "tidak ditemukan akun uji mhs di database\n")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	switch *mode {
	case "crowd":
		var if399ID int
		err := pool.QueryRow(ctx, "SELECT id FROM courses WHERE code = 'IF399'").Scan(&if399ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kursus IF399 tidak ditemukan: %v\n", err)
			os.Exit(1)
		}

		tokens := make([]string, len(testUsers))
		for i, u := range testUsers {
			tok, _ := jwtManager.GenerateAccess(model.User{ID: u.ID, Username: u.Username, Role: "user"})
			tokens[i] = tok
		}

		n := 60
		if len(tokens) < n {
			n = len(tokens)
		}

		fmt.Printf("%d mahasiswa berebut kelas IF399 secara serentak...\n", n)
		t0 := time.Now()
		tally := fire(n, func(i int) string {
			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/courses/%d/enroll", *baseURL, if399ID), bytes.NewReader([]byte("{}")))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tokens[i])
			resp, err := client.Do(req)
			if err != nil {
				return "ERROR: " + err.Error()
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				return "201 OK"
			}
			body, _ := io.ReadAll(resp.Body)
			var errResp model.ErrorResponse
			if json.Unmarshal(body, &errResp) == nil && errResp.Code != "" {
				return fmt.Sprintf("%d %s", resp.StatusCode, errResp.Code)
			}
			return fmt.Sprintf("%d (unknown)", resp.StatusCode)
		})
		elapsed := time.Since(t0)

		fmt.Println("hasil request:")
		for k, v := range tally {
			fmt.Printf("  %-30s %d\n", k, v)
		}
		fmt.Printf("waktu total : %v\n", elapsed)

		var quota, enrolled int
		_ = pool.QueryRow(ctx, "SELECT quota, enrolled_count FROM courses WHERE id = $1", if399ID).Scan(&quota, &enrolled)
		var actualRows int
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if399ID).Scan(&actualRows)

		status := "TIDAK KONSISTEN"
		if enrolled == actualRows && enrolled <= quota {
			status = "KONSISTEN"
		}
		fmt.Printf("IF399 kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n", quota, enrolled, actualRows, status)

	case "credits":
		type CourseMini struct {
			ID      int
			Code    string
			Credits int
		}
		var courses []CourseMini
		cRows, _ := pool.Query(ctx, "SELECT id, code, credits FROM courses WHERE code != 'IF399' ORDER BY id LIMIT 12")
		for cRows.Next() {
			var cm CourseMini
			if err := cRows.Scan(&cm.ID, &cm.Code, &cm.Credits); err == nil {
				courses = append(courses, cm)
			}
		}
		cRows.Close()

		numStudents := 10
		if len(testUsers) < numStudents {
			numStudents = len(testUsers)
		}
		numCourses := len(courses)
		totalReqs := numStudents * numCourses

		type jobDef struct {
			userIndex   int
			courseIndex int
		}
		jobs := make([]jobDef, totalReqs)
		idx := 0
		for s := 0; s < numStudents; s++ {
			for c := 0; c < numCourses; c++ {
				jobs[idx] = jobDef{userIndex: s, courseIndex: c}
				idx++
			}
		}

		tokens := make([]string, numStudents)
		for s := 0; s < numStudents; s++ {
			tok, _ := jwtManager.GenerateAccess(model.User{ID: testUsers[s].ID, Username: testUsers[s].Username, Role: "user"})
			tokens[s] = tok
		}

		fmt.Printf("%d mahasiswa, masing-masing mendaftar ke %d kelas 3 SKS secara serentak (batas 24 SKS)...\n", numStudents, numCourses)
		t0 := time.Now()
		tally := fire(totalReqs, func(i int) string {
			j := jobs[i]
			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/courses/%d/enroll", *baseURL, courses[j.courseIndex].ID), bytes.NewReader([]byte("{}")))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tokens[j.userIndex])
			resp, err := client.Do(req)
			if err != nil {
				return "ERROR: " + err.Error()
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				return "201 OK"
			}
			body, _ := io.ReadAll(resp.Body)
			var errResp model.ErrorResponse
			if json.Unmarshal(body, &errResp) == nil && errResp.Code != "" {
				return fmt.Sprintf("%d %s", resp.StatusCode, errResp.Code)
			}
			return fmt.Sprintf("%d (unknown)", resp.StatusCode)
		})
		elapsed := time.Since(t0)

		fmt.Println("hasil request:")
		for k, v := range tally {
			fmt.Printf("  %-30s %d\n", k, v)
		}
		fmt.Printf("waktu total : %v\n", elapsed)

		allValid := true
		for s := 0; s < numStudents; s++ {
			var totalCredits int
			_ = pool.QueryRow(ctx, `SELECT COALESCE(SUM(c.credits), 0) FROM enrollments e JOIN courses c ON c.id = e.course_id WHERE e.user_id = $1`, testUsers[s].ID).Scan(&totalCredits)
			if totalCredits > 24 {
				allValid = false
			}
		}
		cStatus := "TIDAK KONSISTEN"
		if allValid {
			cStatus = "KONSISTEN"
		}
		fmt.Printf("seluruh %d mahasiswa <= 24 SKS -> %s\n", numStudents, cStatus)

	case "swap":
		var if101ID, if102ID int
		_ = pool.QueryRow(ctx, "SELECT id FROM courses WHERE code = 'IF101'").Scan(&if101ID)
		_ = pool.QueryRow(ctx, "SELECT id FROM courses WHERE code = 'IF102'").Scan(&if102ID)

		for i := 0; i < 20; i++ {
			_, _ = pool.Exec(ctx, "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", if101ID, testUsers[i].ID)
		}
		for i := 20; i < 40; i++ {
			_, _ = pool.Exec(ctx, "INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", if102ID, testUsers[i].ID)
		}
		_, _ = pool.Exec(ctx, "UPDATE courses SET enrolled_count = (SELECT COUNT(*) FROM enrollments WHERE course_id = courses.id) WHERE id IN ($1, $2)", if101ID, if102ID)

		type swapJob struct {
			token string
			from  int
			to    int
		}
		jobs := make([]swapJob, 40)
		for i := 0; i < 20; i++ {
			tok, _ := jwtManager.GenerateAccess(model.User{ID: testUsers[i].ID, Username: testUsers[i].Username, Role: "user"})
			jobs[i] = swapJob{token: tok, from: if101ID, to: if102ID}
		}
		for i := 20; i < 40; i++ {
			tok, _ := jwtManager.GenerateAccess(model.User{ID: testUsers[i].ID, Username: testUsers[i].Username, Role: "user"})
			jobs[i] = swapJob{token: tok, from: if102ID, to: if101ID}
		}

		fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")
		t0 := time.Now()
		tally := fire(40, func(i int) string {
			j := jobs[i]
			payload, _ := json.Marshal(map[string]int{"from_course_id": j.from, "to_course_id": j.to})
			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/enrollments/swap", *baseURL), bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+j.token)
			resp, err := client.Do(req)
			if err != nil {
				return "ERROR: " + err.Error()
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return "200 OK"
			}
			body, _ := io.ReadAll(resp.Body)
			var errResp model.ErrorResponse
			if json.Unmarshal(body, &errResp) == nil && errResp.Code != "" {
				return fmt.Sprintf("%d %s", resp.StatusCode, errResp.Code)
			}
			return fmt.Sprintf("%d (unknown)", resp.StatusCode)
		})
		elapsed := time.Since(t0)

		fmt.Println("hasil request:")
		for k, v := range tally {
			fmt.Printf("  %-30s %d\n", k, v)
		}
		fmt.Printf("waktu total : %v\n", elapsed)

		var c1Quota, c1Enrolled, c1Rows int
		var c2Quota, c2Enrolled, c2Rows int
		_ = pool.QueryRow(ctx, "SELECT quota, enrolled_count FROM courses WHERE id = $1", if101ID).Scan(&c1Quota, &c1Enrolled)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if101ID).Scan(&c1Rows)
		_ = pool.QueryRow(ctx, "SELECT quota, enrolled_count FROM courses WHERE id = $1", if102ID).Scan(&c2Quota, &c2Enrolled)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if102ID).Scan(&c2Rows)

		fmt.Printf("IF101 kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n", c1Quota, c1Enrolled, c1Rows, func() string {
			if c1Enrolled == c1Rows {
				return "KONSISTEN"
			}
			return "TIDAK KONSISTEN"
		}())
		fmt.Printf("IF102 kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n", c2Quota, c2Enrolled, c2Rows, func() string {
			if c2Enrolled == c2Rows {
				return "KONSISTEN"
			}
			return "TIDAK KONSISTEN"
		}())

	case "supervisor":
		adminUser := model.User{ID: 1, Username: "admin", Role: "admin"}
		adminToken, _ := jwtManager.GenerateAccess(adminUser)

		var supID int
		err := pool.QueryRow(ctx, "SELECT id FROM supervisors WHERE id = 1").Scan(&supID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "pembimbing ID 1 tidak ditemukan: %v\n", err)
			os.Exit(1)
		}

		type StudentIDInfo struct {
			ID int
		}
		var students []StudentIDInfo
		sRows, _ := pool.Query(ctx, "SELECT id FROM students ORDER BY id LIMIT 30")
		for sRows.Next() {
			var s StudentIDInfo
			if err := sRows.Scan(&s.ID); err == nil {
				students = append(students, s)
			}
		}
		sRows.Close()

		n := len(students)
		fmt.Printf("%d mahasiswa serentak ditetapkan ke pembimbing (kuota 3)...\n", n)
		t0 := time.Now()
		tally := fire(n, func(i int) string {
			payload, _ := json.Marshal(map[string]int{"student_id": students[i].ID})
			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/supervisors/%d/assign", *baseURL, supID), bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+adminToken)
			resp, err := client.Do(req)
			if err != nil {
				return "ERROR: " + err.Error()
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				return "201 OK"
			}
			body, _ := io.ReadAll(resp.Body)
			var errResp model.ErrorResponse
			if json.Unmarshal(body, &errResp) == nil && errResp.Code != "" {
				return fmt.Sprintf("%d %s", resp.StatusCode, errResp.Code)
			}
			return fmt.Sprintf("%d (unknown)", resp.StatusCode)
		})
		elapsed := time.Since(t0)

		fmt.Println("hasil request:")
		for k, v := range tally {
			fmt.Printf("  %-30s %d\n", k, v)
		}
		fmt.Printf("waktu total : %v\n", elapsed)

		var quota, supervised int
		_ = pool.QueryRow(ctx, "SELECT quota, supervised_count FROM supervisors WHERE id = $1", supID).Scan(&quota, &supervised)
		var actualRows int
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM supervisions WHERE supervisor_id = $1", supID).Scan(&actualRows)

		status := "TIDAK KONSISTEN"
		if supervised == actualRows && supervised <= quota {
			status = "KONSISTEN"
		}
		fmt.Printf("Supervisor %d kuota=%d supervised_count=%d baris supervisions=%d -> %s\n", supID, quota, supervised, actualRows, status)
	}
}

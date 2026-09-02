package service

import (
	"testing"

	"api-students/app/model"
)

// Perhatikan: pengujian ini tidak menyalakan server, tidak menyentuh
// database, dan tidak membuat fiber.Ctx.
func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{137, 20, 7},
	}

	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d",
				tc.total, tc.limit, tc.want, got)
		}
	}
}

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "123", Name: "Sari", Grade: "A", IsActive: true}
	inactive := false

	result, errs := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})

	if len(errs) != 0 {
		t.Fatalf("tidak seharusnya ada error: %v", errs)
	}
	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Name != "Sari" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestValidateCreate(t *testing.T) {
	errs := ValidateCreate(model.CreateStudentRequest{NIM: "", Name: ""})
	if len(errs) != 2 {
		t.Fatalf("harap 2 error, dapat %d", len(errs))
	}
	if errs["nim"] == "" || errs["name"] == "" {
		t.Errorf("pesan error untuk nim/name tidak boleh kosong")
	}
}

func TestValidateReplace(t *testing.T) {
	errs := ValidateReplace(model.ReplaceStudentRequest{NIM: "123", Name: "  "})
	if len(errs) != 1 {
		t.Fatalf("harap 1 error, dapat %d", len(errs))
	}
	if errs["name"] == "" {
		t.Errorf("pesan error untuk name tidak boleh kosong")
	}
}

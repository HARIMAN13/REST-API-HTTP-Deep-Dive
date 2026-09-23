package service

import (
	"testing"

	"api-students/app/model"
)

func TestApplyPatch(t *testing.T) {
	initial := model.Student{ID: 1, NIM: "123", Name: "Sari", Grade: "A", IsActive: true}
	inactive := false

	result := ApplyPatch(initial, model.PatchStudentRequest{IsActive: &inactive})

	if result.IsActive {
		t.Error("is_active seharusnya berubah menjadi false")
	}
	if result.Name != "Sari" {
		t.Error("field yang tidak dikirim seharusnya tidak berubah")
	}
}

func TestIsEmptyPatch(t *testing.T) {
	if !IsEmptyPatch(model.PatchStudentRequest{}) {
		t.Error("struct kosong seharusnya dikenali sebagai empty patch")
	}
	
	active := true
	if IsEmptyPatch(model.PatchStudentRequest{IsActive: &active}) {
		t.Error("struct dengan satu field seharusnya tidak empty")
	}
}

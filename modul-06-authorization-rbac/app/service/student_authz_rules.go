package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang boleh menyentuh data mahasiswa.
// Dua jalur yang diizinkan:
// 1. Kepemilikan (ownership) — data itu miliknya sendiri (ownerID == current.UserID).
// 2. Permission — role-nya memang berhak atas data siapa pun (contoh: student:update:any).
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

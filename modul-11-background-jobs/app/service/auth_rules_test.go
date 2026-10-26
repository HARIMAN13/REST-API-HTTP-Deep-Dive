package service

import "testing"

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{
			name:     "Too short",
			password: "short",
			want:     "minimal 8 karakter",
		},
		{
			name:     "No digits",
			password: "passwordtanpaangka",
			want:     "harus memuat huruf dan angka",
		},
		{
			name:     "No letters",
			password: "1234567890",
			want:     "harus memuat huruf dan angka",
		},
		{
			name:     "Common password",
			password: "password123",
			want:     "password terlalu umum",
		},
		{
			name:     "Valid strong password",
			password: "StrongPassword123!",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPasswordStrength(tt.password); got != tt.want {
				t.Errorf("checkPasswordStrength() = %v, want %v", got, tt.want)
			}
		})
	}
}

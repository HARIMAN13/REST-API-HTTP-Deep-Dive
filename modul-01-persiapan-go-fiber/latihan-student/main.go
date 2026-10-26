package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

func (s Student) GetInfo() {
	fmt.Printf("ID: %d | Nama: %s | Grade: %.2f | Aktif: %t\n", s.ID, s.Name, s.Grade, s.IsActive)
}

func (s *Student) UpdateGrade(newGrade float64) {
	s.Grade = newGrade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	student := Student{
		ID:       1,
		Name:     "Rina",
		Grade:    85.5,
		IsActive: false,
	}

	fmt.Println("=== Informasi Awal ===")
	student.GetInfo()

	fmt.Println("\n=== Update Grade ===")
	student.UpdateGrade(90)
	student.GetInfo()

	fmt.Println("\n=== Activate ===")
	student.Activate()
	student.GetInfo()

	fmt.Println("\n=== Deactivate ===")
	student.Deactivate()
	student.GetInfo()

	fmt.Println("\n=== Penjelasan Receiver ===")
	fmt.Println("- GetInfo() menggunakan value receiver karena hanya membaca data tanpa mengubah state.")
	fmt.Println("- UpdateGrade(), Activate(), Deactivate() menggunakan pointer receiver karena mengubah data asli struct.")
}

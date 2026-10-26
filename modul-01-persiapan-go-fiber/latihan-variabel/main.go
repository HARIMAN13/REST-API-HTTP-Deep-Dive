package main

import "fmt"

func main() {
	name := "Andi"
	age := 20
	height := 175.5
	isStudent := true
	numbers := []int{10, 20, 30}

	students := map[string]int{
		"Ayu": 90,
		"Budi": 80,
		"Cici": 85,
	}

	fmt.Println("=== Variabel Dasar ===")
	fmt.Println("Nama:", name)
	fmt.Println("Umur:", age)
	fmt.Println("Tinggi:", height)
	fmt.Println("Status Mahasiswa:", isStudent)
	fmt.Println("Slice:", numbers)
	fmt.Println()

	fmt.Println("=== Map Mahasiswa ===")
	fmt.Println("Data awal:", students)

	students["Deni"] = 88
	fmt.Println("Setelah tambah Deni:", students)

	value, exists := students["Budi"]
	if exists {
		fmt.Println("Nilai Budi:", value)
	} else {
		fmt.Println("Budi tidak ditemukan")
	}

	delete(students, "Ayu")
	fmt.Println("Setelah hapus Ayu:", students)

	fmt.Println("Semua data mahasiswa:")
	for key, value := range students {
		fmt.Printf("- %s: %d\n", key, value)
	}
}

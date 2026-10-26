package main

import "fmt"

func swap(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func passByValue(x int) {
	x = 99
	fmt.Println("Di dalam passByValue:", x)
}

func passByPointer(x *int) {
	*x = 99
	fmt.Println("Di dalam passByPointer:", *x)
}

func main() {
	fmt.Println("=== Swap dengan Pointer ===")
	a := 10
	b := 20
	fmt.Println("Sebelum swap:", a, b)
	swap(&a, &b)
	fmt.Println("Setelah swap:", a, b)

	fmt.Println("\n=== updateSlice dengan Pointer ===")
	items := []string{"apel", "pisang"}
	fmt.Println("Sebelum update:", items)
	updateSlice(&items, "mangga")
	fmt.Println("Setelah update:", items)

	fmt.Println("\n=== Pass by Value vs Pass by Pointer ===")
	value := 5
	passByValue(value)
	fmt.Println("Setelah passByValue:", value)

	pointerValue := 5
	passByPointer(&pointerValue)
	fmt.Println("Setelah passByPointer:", pointerValue)

	fmt.Println("\nPenjelasan:")
	fmt.Println("- pass by value mengirim salinan data, jadi perubahan di fungsi tidak berubah di luar.")
	fmt.Println("- pass by pointer mengirim alamat memori, jadi perubahan di fungsi akan terlihat di luar.")
}

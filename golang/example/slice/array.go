package slice

import "fmt"

func Add(numbers [5]int) {
	for i := 0; i < len(numbers); i++ {
		numbers[i] = numbers[i] + 1
	}
	fmt.Println("numbers in Add:", numbers) // [2,3,4,5,6]
}

func ArrayExample() {
	var numbers [5]int
	for i := 0; i < len(numbers); i++ {
		numbers[i] = i + 1
	}
	Add(numbers)
	fmt.Println("numbers in main: ", numbers) // [1,2,3,4,5]
}

package main

import "fmt"

func AskForWord(a1 string) {
	var word string
	fmt.Println("give me a word")
	fmt.Scanln(&word)
	fmt.Println(word)
}

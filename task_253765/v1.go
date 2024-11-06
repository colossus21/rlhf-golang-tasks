//go:build v1
// +build v1

package main

import "fmt"

const (
	Apple = iota
	Banana
	Orange
)

var fruitNames = map[string]string{
	"apple":  "Red Delicious",
	"banana": "Cavendish",
}

func GetFruitName(fruit string) string {
	return fruitNames[fruit]
}

func main() {
	enumMap := map[int]string{
		Apple:  "Apple",
		Banana: "Banana",
		Orange: "Orange",
	}

	fmt.Println(enumMap[Apple])
	fmt.Println(GetFruitName("apple"))
}

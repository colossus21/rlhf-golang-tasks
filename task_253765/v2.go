//go:build v2
// +build v2

package main

import "fmt"

var myMap = map[string]int{
	"one":   1,
	"two":   2,
	"three": 3,
}

func GetValue(key string) (int, bool) {
	value, exists := myMap[key]
	return value, exists
}

func main() {
	value, exists := GetValue("one")
	if exists {
		fmt.Println("Value:", value)
	}

	directValue := myMap["two"]
	fmt.Println("Direct value:", directValue)
}

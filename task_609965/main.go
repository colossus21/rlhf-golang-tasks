package main

import "fmt"
import service "rlhf/realestate"

func main() {
	s := service.NewRealEstateSystem()
	s.AddProperty(service.NewProperty(1, 1, 1))
	s.AddProperty(service.NewProperty(2, 2, 1))
	s.AddProperty(service.NewProperty(3, 3, 1))
	s.AddProperty(service.NewProperty(4, 4, 3))
	s.AddProperty(service.NewProperty(5, 5, 2))
	fmt.Println(s.FilterByPriceRange(1, 2))
	fmt.Println(s.FilterBySizePreference(1, 2))
}

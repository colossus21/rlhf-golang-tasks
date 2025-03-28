//go:build v1
// +build v1

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// RealEstate represents a real estate property
type RealEstate struct {
	ID           int    `json:"id"`
	Location     string `json:"location"`
	PropertyType string `json:"propertyType"`
	Price        int    `json:"price"`
}

// realEstates is the mocked data source
var realEstates = []RealEstate{
	{ID: 1, Location: "New York", PropertyType: "Apartment", Price: 1200000},
	{ID: 2, Location: "San Francisco", PropertyType: "House", Price: 1500000},
	{ID: 3, Location: "Los Angeles", PropertyType: "Townhouse", Price: 950000},
	{ID: 4, Location: "Chicago", PropertyType: "Apartment", Price: 700000},
	{ID: 5, Location: "Houston", PropertyType: "House", Price: 500000},
	{ID: 6, Location: "Phoenix", PropertyType: "Apartment", Price: 650000},
	{ID: 7, Location: "Philadelphia", PropertyType: "Townhouse", Price: 450000},
	{ID: 8, Location: "San Antonio", PropertyType: "House", Price: 400000},
	{ID: 9, Location: "San Diego", PropertyType: "Apartment", Price: 850000},
	{ID: 10, Location: "Dallas", PropertyType: "Townhouse", Price: 550000},
}

// searchRealEstate handles the search request
func searchRealEstate(w http.ResponseWriter, r *http.Request) {
	var result []RealEstate

	minPrice, err := strconv.Atoi(r.URL.Query().Get("minPrice"))
	if err != nil {
		minPrice = 0
	}

	maxPrice, err := strconv.Atoi(r.URL.Query().Get("maxPrice"))
	if err != nil {
		maxPrice = 99999999
	}

	location := r.URL.Query().Get("location")
	propertyType := r.URL.Query().Get("propertyType")

	for _, estate := range realEstates {
		if estate.Price >= minPrice && estate.Price <= maxPrice {
			if location == "" || estate.Location == location {
				if propertyType == "" || estate.PropertyType == propertyType {
					result = append(result, estate)
				}
			}
		}
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/search", searchRealEstate)
	fmt.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server error:", err)
	}
}

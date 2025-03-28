//go:build v2
// +build v2

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// RealEstate struct represents a real estate property
type RealEstate struct {
	ID           int     `json:"id"`
	PropertyType string  `json:"property_type"`
	Price        float64 `json:"price"`
	Location     string  `json:"location"`
}

// Mock data source for real estate properties
var realEstates = []RealEstate{
	{ID: 1, PropertyType: "Apartment", Price: 250000.00, Location: "New York City"},
	{ID: 2, PropertyType: "House", Price: 500000.00, Location: "Los Angeles"},
	{ID: 3, PropertyType: "Condo", Price: 300000.00, Location: "San Francisco"},
	{ID: 4, PropertyType: "Townhouse", Price: 400000.00, Location: "Chicago"},
	{ID: 5, PropertyType: "Villa", Price: 700000.00, Location: "Miami"},
	{ID: 6, PropertyType: "Apartment", Price: 180000.00, Location: "Houston"},
	{ID: 7, PropertyType: "House", Price: 600000.00, Location: "Phoenix"},
	{ID: 8, PropertyType: "Condo", Price: 280000.00, Location: "Washington D.C."},
	{ID: 9, PropertyType: "Townhouse", Price: 350000.00, Location: "Philadelphia"},
	{ID: 10, PropertyType: "Apartment", Price: 220000.00, Location: "Seattle"},
}

// searchRealEstates handler function
func searchRealEstates(w http.ResponseWriter, r *http.Request) {
	// Parse URL query parameters
	query := r.URL.Query()
	minPriceStr := query.Get("min_price")
	maxPriceStr := query.Get("max_price")
	propertyType := query.Get("property_type")
	location := query.Get("location")

	// Validate query parameters
	var minPrice, maxPrice float64
	if minPriceStr != "" {
		minPrice, _ = strconv.ParseFloat(minPriceStr, 64)
	}
	if maxPriceStr != "" {
		maxPrice, _ = strconv.ParseFloat(maxPriceStr, 64)
	}

	// Filter real estates based on query parameters
	filteredRealEstates := []RealEstate{}
	for _, estate := range realEstates {
		if (minPrice == 0 || estate.Price >= minPrice) && (maxPrice == 0 || estate.Price <= maxPrice) && (propertyType == "" || estate.PropertyType == propertyType) && (location == "" || strings.Contains(strings.ToLower(estate.Location), strings.ToLower(location))) {
			filteredRealEstates = append(filteredRealEstates, estate)
		}
	}

	// Encode the filtered results as JSON and send as response
	fmt.Fprintf(w, `{"real_estates":%v}`, filteredRealEstates)
}

func main() {
	http.HandleFunc("/search", searchRealEstates)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

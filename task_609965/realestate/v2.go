/*
Package realestate implements a comprehensive property management and filtering system.
This system allows realtors to efficiently manage property listings and provide
tailored property searches for clients through various filtering mechanisms.

Key Features:
- Property management (add, update)
- Advanced filtering capabilities (price, size)
- Status tracking for properties
*/
package realestate

import (
    "errors"
)

// Property represents a real estate property with essential characteristics
type Property struct {
    ID     int     // Unique identifier for the property
    Price  float64 // Price of the property in currency units
    Size   float64 // Size of the property in square feet/meters
    Status string  // Current status of the property (e.g., "available", "sold", "pending")
}

// NewProperty creates and returns a new Property instance with default "available" status
// Parameters:
//   - id: unique identifier for the property
//   - price: listing price of the property
//   - size: total area of the property
// Returns:
//   - *Property: pointer to the newly created Property
func NewProperty(id int, price, size float64) *Property {
    return &Property{
        ID:     id,
        Price:  price,
        Size:   size,
        Status: "available",
    }
}

// RealEstateSystem manages a collection of properties and provides filtering capabilities
type RealEstateSystem struct {
    Properties []*Property // Slice containing all registered properties
}

// NewRealEstateSystem initializes and returns a new RealEstateSystem
// Returns:
//   - *RealEstateSystem: pointer to the newly created system
func NewRealEstateSystem() *RealEstateSystem {
    return &RealEstateSystem{
        Properties: make([]*Property, 0),
    }
}

// AddProperty adds a new property to the system
// Parameters:
//   - property: pointer to the Property to be added
func (res *RealEstateSystem) AddProperty(property *Property) {
    res.Properties = append(res.Properties, property)
}

// UpdatePropertyStatus updates the status of a specific property
// Parameters:
//   - propertyID: ID of the property to update
//   - newStatus: new status to be set
// Returns:
//   - error: nil if successful, error message if property not found
func (res *RealEstateSystem) UpdatePropertyStatus(propertyID int, newStatus string) error {
    for _, property := range res.Properties {
        if property.ID == propertyID {
            property.Status = newStatus
            return nil
        }
    }
    return errors.New("property not found")
}

// FilterByPriceRange returns properties within the specified price range
// Parameters:
//   - minPrice: minimum price threshold
//   - maxPrice: maximum price threshold
// Returns:
//   - []*Property: slice of properties matching the price criteria
func (res *RealEstateSystem) FilterByPriceRange(minPrice, maxPrice float64) []*Property {
    filteredProperties := make([]*Property, 0)
    for _, property := range res.Properties {
        if property.Price >= minPrice && property.Price <= maxPrice {
            filteredProperties = append(filteredProperties, property)
        }
    }
    return filteredProperties
}

// FilterBySizePreference returns properties within the specified size range
// Parameters:
//   - minSize: minimum size threshold
//   - maxSize: maximum size threshold
// Returns:
//   - []*Property: slice of properties matching the size criteria
func (res *RealEstateSystem) FilterBySizePreference(minSize, maxSize float64) []*Property {
    filteredProperties := make([]*Property, 0)
    for _, property := range res.Properties {
        if property.Size >= minSize && property.Size <= maxSize {
            filteredProperties = append(filteredProperties, property)
        }
    }
    return filteredProperties
}
//go:build v2
// +build v2

package main

import (
	"testing"
)

func TestConstantMapDeclaration(t *testing.T) {
	t.Run("attempt constant map declaration", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected compile error for constant map declaration")
			}
		}()

		const testMap = map[string]int{
			"one": 1,
		}
		_ = testMap
	})

	t.Run("verify global map behavior", func(t *testing.T) {
		val, exists := GetValue("one")
		if !exists || val != 1 {
			t.Errorf("Expected 1, got %v with exists=%v", val, exists)
		}
	})
}

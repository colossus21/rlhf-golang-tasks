//go:build v1
// +build v1

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

		const testMap = map[string]string{
			"test": "value",
		}
		_ = testMap
	})

	t.Run("verify regular map works", func(t *testing.T) {
		regularMap := map[string]string{
			"test": "value",
		}
		if regularMap["test"] != "value" {
			t.Errorf("Expected 'value', got %v", regularMap["test"])
		}
	})
}

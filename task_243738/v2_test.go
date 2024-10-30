//go:build v2
// +build v2

package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestXMLToJSONForceList(t *testing.T) {
	var totalTests, passedTests int
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name: "Multiple Parameters",
			input: `<root>
                <item><name>item1</name><value>value1</value></item>
                <item><name>item2</name><value>value2</value></item>
            </root>`,
			expected: `{"Items":[{"Name":"item1","Value":"value1"},{"Name":"item2","Value":"value2"}]}`,
			wantErr:  false,
		},
		{
			name: "Single Parameter",
			input: `<root>
                <item><name>item1</name><value>value1</value></item>
            </root>`,
			expected: `{"Items":[{"Name":"item1","Value":"value1"}]}`,
			wantErr:  false,
		},
		{
			name:     "Empty Parameters",
			input:    `<root></root>`,
			expected: `{"Items":null}`,
			wantErr:  false,
		},
		{
			name: "Nested Parameters",
			input: `<root>
                <item>
                    <name>parent</name>
                    <value><item><name>child</name><value>childValue</value></item></value>
                </item>
            </root>`,
			expected: `{"Items":[{"Name":"parent","Value":"<item><name>child</name><value>childValue</value></item>"}]}`,
			wantErr:  false,
		},
		{
			name: "Special Characters",
			input: `<root>
                <item><name>special&amp;</name><value>&lt;test&gt;</value></item>
            </root>`,
			expected: `{"Items":[{"Name":"special&","Value":"<test>"}]}`,
			wantErr:  false,
		},
		{
			name:    "Invalid XML",
			input:   `<root><item><name>test</root>`,
			wantErr: true,
		},
		{
			name:    "Empty Input",
			input:   "",
			wantErr: true,
		},
	}

	for i, tt := range tests {
		totalTests++
		t.Run(tt.name, func(t *testing.T) {
			result, err := xmlToJSONForceList([]byte(tt.input))

			if tt.wantErr != (err != nil) {
				fmt.Printf("Test %d# %s (Failed)\n", i+1, tt.name)
				t.Fatalf("Expected error: %v, got error: %v", tt.wantErr, err != nil)
				return
			}

			if !tt.wantErr {
				var expectedJSON, resultJSON interface{}
				json.Unmarshal([]byte(tt.expected), &expectedJSON)
				json.Unmarshal(result, &resultJSON)

				if !reflect.DeepEqual(expectedJSON, resultJSON) {
					fmt.Printf("Test %d# %s (Failed)\n", i+1, tt.name)
					t.Errorf("Expected: %s\nGot: %s", tt.expected, string(result))
					return
				}
			}
			passedTests++
			fmt.Printf("Test %d# %s (Passed)\n", i+1, tt.name)
		})
	}
	fmt.Printf("\nTests Passed %d/%d\n", passedTests, totalTests)
}

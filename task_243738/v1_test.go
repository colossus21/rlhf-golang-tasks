//go:build v1
// +build v1

package main

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"testing"
)

func TestXMLToJSONConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name: "Test 1# Multiple Items",
			input: `<root>
  <item>
    <name>item1</name>
    <value>value1</value>
  </item>
  <item>
    <name>item2</name>
    <value>value2</value>
  </item>
</root>`,
			expected: `{
  "Items": [
    {
      "Name": "item1",
      "Value": "value1"
    },
    {
      "Name": "item2",
      "Value": "value2"
    }
  ]
}`,
			wantErr: false,
		},
		{
			name: "Test 2# Single Item Forces List",
			input: `<root>
  <item>
    <name>item1</name>
    <value>value1</value>
  </item>
</root>`,
			expected: `{
  "Items": [
    {
      "Name": "item1",
      "Value": "value1"
    }
  ]
}`,
			wantErr: false,
		},
		{
			name:  "Test 3# Empty Root",
			input: `<root></root>`,
			expected: `{
  "Items": null
}`,
			wantErr: false,
		},
		{
			name: "Test 4# Nested Items",
			input: `<root>
  <item>
    <name>parent</name>
    <value>
      <item>
        <name>child</name>
        <value>childValue</value>
      </item>
    </value>
  </item>
</root>`,
			expected: `{
  "Items": [
    {
      "Name": "parent",
      "Value": "\n      <item>\n        <name>child</name>\n        <value>childValue</value>\n      </item>\n    "
    }
  ]
}`,
			wantErr: false,
		},
		{
			name: "Test 5# Special Characters",
			input: `<root>
  <item>
    <name>special&amp;</name>
    <value>&lt;test&gt;</value>
  </item>
</root>`,
			expected: `{
  "Items": [
    {
      "Name": "special&",
      "Value": "<test>"
    }
  ]
}`,
			wantErr: false,
		},
		{
			name:    "Test 6# Invalid XML",
			input:   `<root><item><name>test</root>`,
			wantErr: true,
		},
		{
			name:    "Test 7# Empty Input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var xmlStruct MyXML
			err := xml.Unmarshal([]byte(tt.input), &xmlStruct)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to unmarshal XML: %v", err)
			}

			result, err := json.MarshalIndent(forceList(xmlStruct), "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal to JSON: %v", err)
			}

			var expectedJSON, resultJSON interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expectedJSON); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}
			if err := json.Unmarshal(result, &resultJSON); err != nil {
				t.Fatalf("Failed to unmarshal result JSON: %v", err)
			}

			if !reflect.DeepEqual(expectedJSON, resultJSON) {
				t.Errorf("\nExpected:\n%s\nGot:\n%s", tt.expected, string(result))
			}
		})
	}
}

//go:build v1
// +build v1

package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// Define XML structure
type MyXML struct {
	XMLName xml.Name `xml:"root"`
	Items   []Item   `xml:"item"`
}

type Item struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

func main() {
	// Sample XML data
	xmlData := `
<root>
  <item>
    <name>item1</name>
    <value>value1</value>
  </item>
  <item>
    <name>item2</name>
    <value>value2</value>
  </item>
</root>
`

	// Unmarshal XML data
	var xmlStruct MyXML
	err := xml.Unmarshal([]byte(xmlData), &xmlStruct)
	if err != nil {
		panic(err)
	}

	// Convert to JSON with forced list for "items"
	jsonData, err := json.MarshalIndent(forceList(xmlStruct), "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))
}

func forceList(v interface{}) interface{} {
	switch v := v.(type) {
	case MyXML:
		// Ensure "items" is always a list
		if len(v.Items) == 1 {
			v.Items = []Item{v.Items[0]}
		}
		return v
	case map[string]interface{}:
		for k, v := range v {
			v[k] = forceList(v)
		}
		return v
	case []interface{}:
		for i, v := range v {
			v[i] = forceList(v)
		}
		return v
	default:
		return v
	}
}

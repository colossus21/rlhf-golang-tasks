//go:build v2
// +build v2

package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// XML struct with a field that can be a single item or a list
type MyXML struct {
	XMLName xml.Name `xml:"root"`
	Items   []Item   `xml:"item"`
}

type Item struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

func main() {
	xmlData := []byte(`
		<root>
			<item>
				<name>item1</name>
				<value>value1</value>
			</item>
		</root>`)

	jsonData, err := xmlToJSONForceList(xmlData)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))

	xmlData = []byte(`
		<root>
			<item>
				<name>item1</name>
				<value>value1</value>
			</item>
			<item>
				<name>item2</name>
				<value>value2</value>
			</item>
		</root>`)

	jsonData, err = xmlToJSONForceList(xmlData)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))
}

// xmlToJSONForceList converts XML to JSON and forces a specific field to be a list
func xmlToJSONForceList(xmlData []byte) ([]byte, error) {
	var data MyXML
	err := xml.Unmarshal(xmlData, &data)
	if err != nil {
		return nil, err
	}

	// Force 'Items' to be a list even if it's a single item
	if len(data.Items) == 1 {
		data.Items = []Item{data.Items[0]}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	// Remove the XMLName field from the JSON output
	jsonData, err = removeXMLName(jsonData)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// removeXMLName removes the XMLName field from the JSON output
func removeXMLName(jsonData []byte) ([]byte, error) {
	var v interface{}
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.UseNumber()
	err := decoder.Decode(&v)
	if err != nil {
		return nil, err
	}

	m, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure")
	}

	delete(m, "XMLName")

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(m)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

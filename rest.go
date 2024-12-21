package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Creature struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"img"`
}

func GetAllEntities() []Creature {
	resp, err := http.Get("https://lovecraftapirest.fly.dev/api/creatures")
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return []Creature{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return []Creature{}
	}

	var creatures []Creature
	err = json.Unmarshal(body, &creatures)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return []Creature{}
	}

	return creatures
}

// UnmarshalJSON customizes the JSON unmarshalling for Creature.
// --> We do not need to modify the code above! The json.Unmarshal function will use this method automatically.
func (c *Creature) UnmarshalJSON(data []byte) error {
	// Define a temporary structure to match the JSON structure,
	// where "img" is an array of strings.
	type Alias Creature
	aux := &struct {
		Img []string `json:"img"` // Temporary field to capture "img" array
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	// Unmarshal the JSON into the temporary structure
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Assign the Image field based on the "img" array
	if len(aux.Img) > 0 {
		c.Image = aux.Img[0] // Take the first image if available
	} else {
		c.Image = "No picture found" // Default message if no images
	}

	return nil
}

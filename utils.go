package main

import (
	"fmt"
	"io/ioutil"

	colorful "github.com/lucasb-eyer/go-colorful"
	yaml "gopkg.in/yaml.v2"
)

// color is a simple wrapper around colorful.Color which adds
// UnmarshalYAML
type color struct {
	colorful.Color
}

// UnmarshalYAML implements yaml.Unmarshaler
func (c *color) UnmarshalYAML(f func(interface{}) error) error {
	var in string
	err := f(&in)
	if err != nil {
		return err
	}

	c.Color, err = colorful.Hex("#" + in)
	return err
}

func readSourcesList(fileName string) (yaml.MapSlice, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var sources yaml.MapSlice
	err = yaml.Unmarshal(data, &sources)
	if err != nil {
		return nil, err
	}

	err = validateMapSlice(sources)
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func validateMapSlice(sources yaml.MapSlice) error {
	// Run through all the values and sanitize them
	dupeSet := make(map[string]struct{})
	for _, item := range sources {
		key, ok := item.Key.(string)
		if !ok {
			return fmt.Errorf("Failed to decode key %q as string", item.Key)
		}

		if _, ok := item.Value.(string); !ok {
			return fmt.Errorf("Failed to decode value %q for key %q as string", item.Value, item.Key)
		}

		if _, ok := dupeSet[key]; ok {
			return fmt.Errorf("Duplicate key %q", key)
		}

		dupeSet[key] = struct{}{}
	}

	return nil
}
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

func buildIcons(assets fs.FS) ([]byte, error) {
	entries, err := fs.ReadDir(assets, ".")
	if err != nil {
		return nil, err
	}

	icons := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := fs.ReadFile(assets, entry.Name())
		if err != nil {
			return nil, err
		}

		name := strings.TrimSuffix(strings.ToLower(entry.Name()), ".svg")
		icons[name] = string(data)
	}

	return json.Marshal(icons)
}

func main() {
	data, err := buildIcons(os.DirFS("assets"))
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("api/icons.json", data, 0644); err != nil {
		panic(err)
	}

	fmt.Println("Icon data saved to api/icons.json")
}

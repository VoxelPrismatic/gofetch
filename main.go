package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/voxelprismatic/gofetch/gofetch"
)

func main() {
	env := map[string]string{}
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		env[parts[0]] = parts[1]
	}
	config := flag.String("config", "/home/"+env["USER"]+"/.config/gofetch.json", "Custom config")
	flag.Parse()
	data, err := os.ReadFile(*config)
	if err != nil {
		panic(err)
	}

	var schema gofetch.Schema
	err = json.Unmarshal(data, &schema)
	if err != nil {
		panic(err)
	}

	lines := schema.Parse()
	for _, line := range lines {
		fmt.Println(line)
	}
}

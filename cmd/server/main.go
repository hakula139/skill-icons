package main

import (
	"log"

	handler "github.com/lelouchfr/skill-icons/api"
)

func main() {
	if err := handler.Run(); err != nil {
		log.Fatal(err)
	}
}

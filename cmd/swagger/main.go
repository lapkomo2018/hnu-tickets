// Command swagger regenerates the OpenAPI specification in docs/ from the
// annotations and DTO types of this module.
package main

import (
	"log"

	"hnu-tickets/internal/platform/apidocs"
)

func main() {
	if _, err := apidocs.Generate(); err != nil {
		log.Fatalf("generate specification: %v", err)
	}
}

package main

import (
	"log"
)

func main() {
	r, err := newRouter()
	if err != nil {
		log.Fatalf("cannot create router: %v", err)
	}

	if err := r.Run(":7564"); err != nil {
		log.Fatal(err)
	}

}

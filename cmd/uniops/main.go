package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/neko233/uniops/internal/server"
)

func main() {
	router := server.NewRouter()

	addr := ":8080"
	fmt.Printf("UniOps server starting on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

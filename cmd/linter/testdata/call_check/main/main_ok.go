package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("error")
	os.Exit(1)
}

func otherFunction() {
	log.Fatal("error")
	os.Exit(1)
}

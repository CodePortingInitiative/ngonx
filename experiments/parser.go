package main

import (
	"log"
	"ngonx/lib/parsers/nginx"
	"os"
)

func main() {
	cwd, _ := os.Getwd()
	filePath := cwd + "/configs/example/nginx.conf"

	log.Println("Parsing NGINX config file:", filePath)

	conf, err := nginx.ParseConfig(filePath)
	if err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	log.Println("Parsed NGINX config successfully")

	// Print the parsed structure
	conf.PrintTree()
}

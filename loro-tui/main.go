package main

import (
	"flag"
	"fmt"
	"log"
	"loro-tui/internal"

	"os"
)

var logger *log.Logger

func main() {
	// write log to a file
	file, err := os.Create("debug.log")
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	// Create a optional flag to active the debug mode
	debug := flag.Bool("debug", true, "Enable debug mode")
	if *debug {
		logger = log.New(file, "", log.LstdFlags)
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	url := flag.String("server", "", "Chat server url (required)")
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: The -server flag is required")
		flag.Usage()
		os.Exit(1)
	}

	loro, err := internal.NewLoro(*url, logger)
	if err != nil {
		fmt.Println("Server: " + err.Error())
		os.Exit(1)
	}

	if err := loro.SetRoot(internal.Pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

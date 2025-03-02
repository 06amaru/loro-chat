package main

import (
	"flag"
	"fmt"
	"loro-tui/internal"
	"os"
)

func main() {
	url := flag.String("server", "", "Chat server url (required)")
	flag.Parse()

	if *url == "" {
		fmt.Println("Error: The -server flag is required")
		flag.Usage()
		os.Exit(1)
	}

	loro, err := internal.NewLoro(*url)
	if err != nil {
		fmt.Println("Server: " + err.Error())
		os.Exit(1)
	}

	if err := loro.SetRoot(internal.Pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

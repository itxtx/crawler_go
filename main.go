package main

import (
	"fmt"
	"os"
)

func main() {
	// Check the number of arguments
	if len(os.Args) < 2 {
		fmt.Println("no website provided")
		os.Exit(1)
	} else if len(os.Args) > 2 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	} else {
		// Get the base URL from the command line arguments
		baseURL := os.Args[1]
		fmt.Printf("starting crawl of: %s\n", baseURL)
		// Here you can call your crawler function
		// crawl(baseURL)
	}
}

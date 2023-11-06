package main

import (
	"log"

	"github.com/mikrolite/mikrolite/internal/commands"
)

func main() {
	rootCmd := commands.NewRoot()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

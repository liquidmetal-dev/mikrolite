package main

import (
	"log"
	"runtime"

	"github.com/mikrolite/mikrolite/internal/commands"
)

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	rootCmd := commands.NewRoot()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

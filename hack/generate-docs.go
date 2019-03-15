package main

import (
	"log"

	"github.com/spf13/cobra/doc"
	"github.com/k14s/kwt/pkg/kwt/cmd"
)

func main() {
	rootCmd := cmd.NewDefaultKwtCmd(nil)

	err := doc.GenMarkdownTree(rootCmd, "./docs/cmd/")
	if err != nil {
		log.Fatal(err)
	}
}

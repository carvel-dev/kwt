package main

import (
	"log"

	"github.com/carvel-dev/kwt/pkg/kwt/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	rootCmd := cmd.NewDefaultKwtCmd(nil)

	err := doc.GenMarkdownTree(rootCmd, "./docs/cmd/")
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"os"

	"github.com/namjeonggil/tc-flowfix-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
package main

import (
	"os"

	"github.com/twotwobread/learning-wiki/internal/cli"
)

// version can be set at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], version, os.Stdout, os.Stderr))
}

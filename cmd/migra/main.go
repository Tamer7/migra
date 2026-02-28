package main

import (
	"github.com/migra/migra/internal/cli"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	cli.Execute()
}

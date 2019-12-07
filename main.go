package main

import (
	"os"

	"github.com/vatriathlon/stopwatch/command"
)

func main() {
	os.Exit(command.Run(os.Args[1:]))
}

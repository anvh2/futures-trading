package main

import (
	cmd "github.com/anvh2/futures-trading/cmd"
)

const (
	version = "0.1.0"
)

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}

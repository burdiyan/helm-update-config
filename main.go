package main

import (
	"os"
)

func main() {
	cmd := newUpdatecfgCmd(nil)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

}

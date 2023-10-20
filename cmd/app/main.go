package main

import (
	"os"

	"lbc/internal/app"
)

func main() {
	if e := app.ReadCommand(); e != nil {
		os.Exit(1)
	}
}

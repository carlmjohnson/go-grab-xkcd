package main

import (
	"os"

	"github.com/erybz/go-grab-xkcd/grabxkcd"
)

func main() {
	os.Exit(grabxkcd.CLI(os.Args[1:]))
}

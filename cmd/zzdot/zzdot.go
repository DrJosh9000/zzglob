package main

import (
	"fmt"
	"os"

	"github.com/DrJosh9000/zzglob"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s pattern\n", os.Args[0])
		os.Exit(1)
	}

	p, err := zzglob.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse pattern %q: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	if err := p.WriteDot(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't write Dot output: %v\n", err)
		os.Exit(1)
	}
}

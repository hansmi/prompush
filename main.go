package main

import (
	"context"
	"flag"
	"log"
)

func main() {
	var opts programOptions

	if err := opts.registerFlags(flag.CommandLine, nil); err != nil {
		log.Fatalf("Error: %v", err)
	}

	flag.Parse()

	p, err := newProgram(opts)

	if err == nil {
		err = p.run(context.Background())
	}

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

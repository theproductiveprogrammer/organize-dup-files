package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type args struct {
	src string
}

func main() {
	args := args{}
	loadArgs(&args)

	f_, err := filepath.Glob(args.src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, fname := range f_ {
		fmt.Println(fname)
	}
}

func loadArgs(args *args) {
	flag.StringVar(&args.src, "src", "*.*", "the source path (glob patterns allowed)")

	flag.Parse()
}

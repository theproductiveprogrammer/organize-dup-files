package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

type args struct {
	Src string `short:"s" long:"src" default:"*.*" description:"the source path (glob patterns allowed)"`
}

func main() {
	args := args{}
	_, err := flags.Parse(&args)
	if err != nil {
		os.Exit(1)
	}

	f_, err := filepath.Glob(args.Src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, fname := range f_ {
		fmt.Println(fname)
	}
}

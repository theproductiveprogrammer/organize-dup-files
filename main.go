package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
)

type args struct {
	Src []string `short:"s" long:"src" default:"*" default:"*/**/*" description:"the source path (glob patterns allowed)"`
	Dst string   `short:"d" long:"dst" default:"." description:"the destination folder (default .)"`
}

func main() {
	args := args{}
	_, err := flags.Parse(&args)
	if err != nil {
		os.Exit(1)
	}

	f_ := []string{}
	for _, src := range args.Src {
		f__, err := filepath.Glob(src)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		f_ = append(f_, f__...)
	}

	for _, fname := range f_ {
		fmt.Println(fname)
	}
}

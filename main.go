package main

import (
	"crypto/sha256"
	"fmt"
	"io"
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
		err := mv(fname, args.Dst)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}

func mv(fname string, dst string) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Mode().IsRegular() {

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		fmt.Printf("%x\t%s\t%s\n", h.Sum(nil), fname, dst)
	}

	return nil
}

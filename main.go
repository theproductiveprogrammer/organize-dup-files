package main

import (
	"crypto/sha256"
	"encoding/hex"
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

	for _, fpath := range f_ {
		err := mv(fpath, args.Dst)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}

func mv(fpath string, dst string) error {
	f, err := os.Open(fpath)
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
		sha := hex.EncodeToString(h.Sum(nil))
		name := filepath.Base(fpath)

		if len(name) > 32 {
			ext := filepath.Ext(fpath)
			sz := 32 - len(ext)
			if sz < 0 {
				ext = ext[0:32]
				sz = 0
			}
			name = name[0:sz] + ext
		}
		d_ := filepath.Join(dst, sha[0:2], sha+"__"+name)
		if d_ != fpath {
			fmt.Printf("%s\t%s\n", fpath, d_)
		}
	}

	return nil
}

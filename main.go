package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jessevdk/go-flags"
)

type args struct {
	Src string   `short:"s" long:"src" default:"." description:"the source folder/file (default .)"`
	Dst string   `short:"d" long:"dst" default:"." description:"the destination folder (default .)"`
	Ext []string `short:"e" long:"ext" description:"a list of file extensions to consider"`
}

/*    way/
 * if not given any extensions, walk the source and list out the extensions,
 * otherwise merge the matching files into the destination
 */
func main() {
	args := args{}
	_, err := flags.Parse(&args)
	if err != nil {
		os.Exit(1)
	}
	if len(args.Ext) == 0 {
		err = listExts(args)
	} else {
		err = mergeMatchingFiles(args)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

/*  way/
 * walk through the files, collect the extensions,
 * and show them to the user
 */
func listExts(args args) error {
	exts := make(map[string]bool)

	err := filepath.WalkDir(args.Src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := filepath.Ext(path)
			if len(ext) > 0 {
				exts[ext] = true
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(exts) == 0 {
		return errors.New("Found no filename extensions in " + args.Src)
	}

	fmt.Printf("Found the following extensions (%s):\n\n", args.Src)

	var ext_ string
	i := 0
	for ext, _ := range exts {
		if len(ext) == 0 {
			continue
		}
		if i != 0 && i%8 == 0 {
			fmt.Println()
		}
		fmt.Printf("%s,", ext)
		i++
		ext_ = ext
	}
	if i%8 != 0 {
		fmt.Println()
	}

	fmt.Printf("\nSelect one or more to organize (e.g.: -e \"%s\")\n", ext_)

	return nil
}

/*  way/
 * Load matching files, and merge them
 * with files existing in the destination to
 * describe how the sources move into the destination
 */
func mergeMatchingFiles(args args) error {
	rules := []rule{}

	err := loadSrcs(args, &rules)
	if err != nil {
		return err
	}

	err = mergeDst(args, &rules)
	if err != nil {
		return err
	}

	describe(rules)

	return nil
}

type rule struct {
	orig string
	sha  string

	clean_name string
}

func loadSrcs(args args, rules *[]rule) error {
	f__, err := filepath.Glob(args.Src)
	if err != nil {
		return err
	}
	fmt.Println(f__)
	/*
		for _, f := range f__ {
			err := loadF(args.Dst, f, rules)
			if err != nil {
				return err
			}
		}
	*/
	return nil
}

func mergeDst(args args, rules *[]rule) error {
	return nil
}
func describe(rules []rule) {
	for _, rule := range rules {
		fmt.Printf("%+v\n", rule)
	}
}

func loadF(dst string, fpath string, rules *[]rule) error {
	for _, r := range *rules {
		if r.orig == fpath {
			return nil
		}
	}

	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return nil
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	*rules = append(*rules, rule{
		orig:       fpath,
		sha:        hex.EncodeToString(h.Sum(nil)),
		clean_name: clean_1(filepath.Base(fpath)),
	})

	return nil
}

func clean_1(n string) string {
	m := regexp.MustCompile(`[^.A-Za-z0-9]+`)
	name := m.ReplaceAllString(n, "_")

	if len(name) > 32 {
		ext := filepath.Ext(n)
		sz := 32 - len(ext)
		if sz < 0 {
			ext = ext[0:32]
			sz = 0
		}
		name = name[0:sz] + ext
	}
	return name
}

func mv(fpath string, dst string) error {
	//rule.clean_name = filepath.Join(dst, rule.sha[0:2], rule.sha+"__"+name)
	return nil
}

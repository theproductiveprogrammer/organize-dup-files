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
	"sort"
	"strings"

	"github.com/jessevdk/go-flags"
)

type args struct {
	Src string   `short:"s" long:"src" default:"." description:"the source folder/file"`
	Dst string   `short:"d" long:"dst" default:"." description:"the destination folder"`
	Ext []string `short:"e" long:"ext" description:"a list of file extensions to consider"`
	Psv bool     `long:"preserve-file-names" description:"if provided, preserve the source filename (default truncates/clean them)"`
}

type srcInfo struct {
	path string
	sha  string

	clean_name string
	dst_ndx    int

	todo string
}

type ByTodo []srcInfo

func (a ByTodo) Len() int           { return len(a) }
func (a ByTodo) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTodo) Less(i, j int) bool { return as_num_1(a[i].todo) < as_num_1(a[j].todo) }
func as_num_1(a string) int {
	i := 0
	if a == "keep" {
		i = 1
	}
	if a == "move" {
		i = 2
	}
	if a == "rmrf" {
		i = 3
	}
	return i
}

type dstInfo struct {
	path string
	sha  string
}

type orgF struct {
	src_f string
	dst_f string
	ext_s []string
	src_i []srcInfo
	dst_i []dstInfo
	mkdir []string
	clean bool
}

/*    way/
 * if not given any extensions, walk the source and list out the extensions,
 * otherwise merge the matching files into the destination
 */
func main() {
	var err error
	args := loadUserArgs()
	if len(args.Ext) == 0 {
		err = listExts(args)
	} else {
		orgf := orgF{
			src_f: args.Src,
			dst_f: args.Dst,
			ext_s: args.Ext,
			src_i: []srcInfo{},
			dst_i: []dstInfo{},
			mkdir: []string{},
			clean: !args.Psv,
		}
		err = mergeMatchingFiles(orgf)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func loadUserArgs() args {
	args := args{}
	_, err := flags.Parse(&args)
	if err != nil {
		os.Exit(1)
	}
	exts := args.Ext
	args.Ext = []string{}
	for _, ext := range exts {
		if ext[0] != '.' {
			ext = "." + ext
		}
		args.Ext = append(args.Ext, ext)
	}
	return args
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
func mergeMatchingFiles(orgf orgF) error {

	err := loadSrcs(&orgf)
	if err != nil {
		return err
	}

	err = mergeDst(&orgf)
	if err != nil {
		return err
	}

	err = describe(orgf)
	if err != nil {
		return err
	}

	return nil
}

func loadSrcs(orgf *orgF) error {
	return filepath.WalkDir(orgf.src_f, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if extMatches(orgf.ext_s, filepath.Ext(path)) {
				err := loadSrc(path, orgf)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func extMatches(exts []string, ext string) bool {
	for _, ext_ := range exts {
		if ext_ == ext {
			return true
		}
	}
	return false
}

/*    way/
 * look for existing matching files in the destination
 * or create a new entry for such a file
 */
func mergeDst(orgf *orgF) error {
	for i, _ := range orgf.src_i {
		err := mergeDst_1(orgf, &orgf.src_i[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func isSameFile(f1, f2 string) bool {
	if f1 == f2 {
		return true
	}
	p1, e1 := filepath.Abs(f1)
	p2, e2 := filepath.Abs(f2)
	return e1 == nil && e2 == nil && p1 == p2
}

func find_in_memory_1(src *srcInfo, dst_i []dstInfo) int {
	for i, dst := range dst_i {
		if dst.sha == src.sha {
			return i
		}
	}
	return -1
}

func find_on_disk_1(src *srcInfo, dstf string, mkdir *[]string) (string, error) {
	var found string
	pfx := src.sha + "__"
	err := filepath.WalkDir(dstf, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				for _, dir := range *mkdir {
					if dir == dstf {
						return nil
					}
				}
				*mkdir = append(*mkdir, dstf)
				return nil
			}
			return err
		}
		if path == dstf {
			return nil
		}
		if d.IsDir() {
			return filepath.SkipDir
		}
		if strings.HasPrefix(filepath.Base(path), pfx) {
			found = path
		}
		return nil
	})

	if len(found) > 0 {
		sha, err := shasum(found)
		if err != nil {
			return "", err
		}
		if sha != src.sha {
			return "", errors.New("File " + found + " has a non-matching sha (" + sha + ")")
		}
	}

	return found, err
}

/*    way/
 * find an existing destination match for the source
 * (in memory or on disk) or create a new destination
 * entry, then set the appropriate action (keep, move, rmrf)
 */
func mergeDst_1(orgf *orgF, src *srcInfo) error {

	i := find_in_memory_1(src, orgf.dst_i)
	if i >= 0 {
		src.dst_ndx = i
		if isSameFile(orgf.dst_i[i].path, src.path) {
			src.todo = "keep"
		} else {
			src.todo = "rmrf"
		}
		return nil
	}

	dstf := filepath.Join(orgf.dst_f, src.sha[0:2])
	found, err := find_on_disk_1(src, dstf, &orgf.mkdir)
	if err != nil {
		return err
	}

	src.dst_ndx = len(orgf.dst_i)
	if len(found) > 0 {

		orgf.dst_i = append(orgf.dst_i, dstInfo{
			path: found,
			sha:  src.sha,
		})

		if isSameFile(found, src.path) {
			src.todo = "keep"
		} else {
			src.todo = "rmrf"
		}

	} else {

		orgf.dst_i = append(orgf.dst_i, dstInfo{
			path: filepath.Join(dstf, src.sha+"__"+src.clean_name),
			sha:  src.sha,
		})

		src.todo = "move"

	}

	return nil

}

/*    way/
 * walk the source files and describe what needs to happen to each of
 * them. Also describe the new directories that need to be created.
 */
func describe(orgf orgF) error {
	for _, fname := range orgf.mkdir {
		fmt.Printf("mkdir %s\n", fname)
	}

	sort.Sort(ByTodo(orgf.src_i))
	for _, inf := range orgf.src_i {
		if inf.todo == "keep" {
			continue
		} else if inf.todo == "move" {
			fmt.Printf("mv %s\t%s\n", shellName(inf.path), shellName(orgf.dst_i[inf.dst_ndx].path))
		} else if inf.todo == "rmrf" {
			fmt.Printf("rm %s\t# %s\n", shellName(inf.path), shellName(orgf.dst_i[inf.dst_ndx].path))
		} else {
			return errors.New("UNEXPECTED ERROR 3253: Did not understand status: " + inf.todo)
		}
	}
	return nil
}

func shellName(s string) string {
	return "'" + strings.Join(strings.Split(s, "'"), `'"'"'`) + "'"
}

func loadSrc(fpath string, orgf *orgF) error {
	sha, err := shasum(fpath)
	if err != nil {
		return err
	}

	var clean_name string
	if orgf.clean {
		clean_name = clean_1(filepath.Base(fpath))
	} else {
		clean_name = filepath.Base(fpath)
	}

	orgf.src_i = append(orgf.src_i, srcInfo{
		path:       fpath,
		sha:        sha,
		clean_name: clean_name,
	})

	return nil
}

func shasum(fpath string) (string, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

/*  understand/
 * because we have a unique sha as the name, we only need
 * to keep what we think is valid text to give it more context
 */
var m1 *regexp.Regexp = regexp.MustCompile(`[^A-Za-z0-9*~!@#$%^&*]+`)
var m2 *regexp.Regexp = regexp.MustCompile(`^.*?[A-Za-z][A-Za-z][A-Za-z]+`)

func clean_1(n string) string {
	ext := filepath.Ext(n)
	n = n[:len(n)-len(ext)]
	name := m1.ReplaceAllString(n, "_")

	s := strings.Split(name, "_")
	r := []string{}
	for _, s_ := range s {
		s_ = m2.FindString(s_)
		if len(s_) > 0 {
			r = append(r, s_)
		}
	}

	name = strings.Join(r, "_")

	return resize_1(name, ext)
}

/*    problem/
 * we want the file size (name + ext) to be less than 32 characters
 *
 *    understand/
 * options
 *    a_really_long_name_with_no_extension
 *    a_really_long_name_with.extension
 *    a_really_long_name_with.a_really_long_extension
 *    a_name.with_a_really_long_extension
 *    .a_really_long_extension_with_no_name
 *
 *    way/
 * truncate the name and keep the extension (as long as we have at least 8 characters
 * the original name)
 */
func resize_1(name, ext string) string {
	sz := len(name)
	sz_e := len(ext)

	if sz+sz_e <= 32 {
		return name + ext
	}

	if sz == 0 {
		return ext[0:32]
	}

	if sz_e == 0 {
		return name[0:32]
	}

	nsz := 32 - sz_e

	if nsz < 8 {
		if sz < 8 {
			sz_e = 32 - sz
			nsz = sz
		} else {
			sz_e = 24
			nsz = 8
		}
	}

	return name[0:nsz] + ext[0:sz_e]
}

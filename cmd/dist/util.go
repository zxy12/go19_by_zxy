package main

import (
	_ "bytes"
	"fmt"
	_ "io"
	_ "io/ioutil"
	"os"
	_ "os/exec"
	"path/filepath"
	_ "runtime"
	"sort"
	_ "strconv"
	"strings"
	"sync"
	_ "time"
)

// pathf is fmt.Sprintf for generating paths
// (on windows it turns / into \ after the printf).
func pathf(format string, args ...interface{}) string {
	return filepath.Clean(fmt.Sprintf(format, args...))
}

// filter returns a slice containing the elements x from list for which f(x) == true.
func filter(list []string, f func(string) bool) []string {
	var out []string
	for _, x := range list {
		if f(x) {
			out = append(out, x)
		}
	}
	return out
}

// uniq returns a sorted slice containing the unique elements of list.
func uniq(list []string) []string {
	out := make([]string, len(list))
	copy(out, list)
	sort.Strings(out)
	keep := out[:0]
	for _, x := range out {
		if len(keep) == 0 || keep[len(keep)-1] != x {
			keep = append(keep, x)
		}
	}
	return keep
}

// splitlines returns a slice with the result of splitting
// the input p after each \n.
func splitlines(p string) []string {
	return strings.SplitAfter(p, "\n")
}

// splitfields replaces the vector v with the result of splitting
// the input p into non-empty fields containing no spaces.
func splitfields(p string) []string {
	return strings.Fields(p)
}

const (
	CheckExit = 1 << iota
	ShowOutput
	Background
)

var outputLock sync.Mutex

func main() {
	os.Setenv("TERM", "dumb") // disable escape codes in clang errors

	// provide -check-armv6k first, before checking for $GOROOT so that
	// it is possible to run this check without having $GOROOT available.
	if len(os.Args) > 1 && os.Args[1] == "-check-armv6k" {

		// 找不到这个方法？为啥
		useARMv6K() // might fail with SIGILL
		println("ARMv6K supported.")
		os.Exit(0)
	}
}

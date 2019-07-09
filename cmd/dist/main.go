package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// count is a flag.Value that is like a flag.Bool and a flag.Int.
// If used as -name, it increments the count, but -name=x sets the count.
// Used for verbose flag -v.
type count int

// cmdtab records the available commands.
var cmdtab = []struct {
	name string
	f    func()
}{

	{"version", cmdversion},

	/*
		{"banner", cmdbanner},
		{"bootstrap", cmdbootstrap},
		{"clean", cmdclean},
		{"env", cmdenv},
		{"install", cmdinstall},
		{"list", cmdlist},
		{"test", cmdtest},
		{"version", cmdversion},
	*/
}

var _p_open bool = false

func (c *count) String() string {
	return fmt.Sprint(int(*c))
}

func (c *count) Set(s string) error {
	switch s {
	case "true":
		*c++
	case "false":
		*c = 0
	default:
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid count %q", s)
		}
		*c = count(n)
	}
	return nil
}

func (c *count) IsBoolFlag() bool {
	return true
}

// main---------- start ---------

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
	slash = string(filepath.Separator)
	_p("slash=" + slash)
	gohostos = runtime.GOOS
	_p("gohostos=", gohostos)

	switch gohostos {
	case "darwin":
		// Even on 64-bit platform, darwin uname -m prints i386.
		// We don't support any of the OS X versions that run on 32-bit-only hardware anymore.
		gohostarch = "amd64"
	case "freebsd":
		// Since FreeBSD 10 gcc is no longer part of the base system.
		defaultclang = true
	case "solaris":
		// Even on 64-bit platform, solaris uname -m prints i86pc.
		out := run("", CheckExit, "isainfo", "-n")
		if strings.Contains(out, "amd64") {
			gohostarch = "amd64"
		}
		if strings.Contains(out, "i386") {
			gohostarch = "386"
		}
	case "plan9":
		gohostarch = os.Getenv("objtype")
		if gohostarch == "" {
			fatal("$objtype is unset")
		}
	case "windows":
		exe = ".exe"
	}

	sysinit()
	_p("gohostarch=" + gohostarch)
	out := run("", CheckExit, "uname", "-m")
	_p("out=" + out)
	if gohostarch == "" {
		// Default Unix system.
		out := run("", CheckExit, "uname", "-m")
		switch {
		case strings.Contains(out, "x86_64"), strings.Contains(out, "amd64"):
			gohostarch = "amd64"
		case strings.Contains(out, "86"):
			gohostarch = "386"
		case strings.Contains(out, "arm"):
			gohostarch = "arm"
		case strings.Contains(out, "aarch64"):
			gohostarch = "arm64"
		case strings.Contains(out, "ppc64le"):
			gohostarch = "ppc64le"
		case strings.Contains(out, "ppc64"):
			gohostarch = "ppc64"
		case strings.Contains(out, "mips64"):
			gohostarch = "mips64"
			if elfIsLittleEndian(os.Args[0]) {
				gohostarch = "mips64le"
			}
		case strings.Contains(out, "mips"):
			gohostarch = "mips"
			if elfIsLittleEndian(os.Args[0]) {
				gohostarch = "mipsle"
			}
		case strings.Contains(out, "s390x"):
			gohostarch = "s390x"
		case gohostos == "darwin":
			if strings.Contains(run("", CheckExit, "uname", "-v"), "RELEASE_ARM_") {
				gohostarch = "arm"
			}
		default:
			fatal("unknown architecture: %s", out)
		}
	}
	if gohostarch == "arm" || gohostarch == "mips64" || gohostarch == "mips64le" {
		maxbg = min(maxbg, runtime.NumCPU())
	}
	bginit()

	// The OS X 10.6 linker does not support external linking mode.
	// See golang.org/issue/5130.
	//
	// OS X 10.6 does not work with clang either, but OS X 10.9 requires it.
	// It seems to work with OS X 10.8, so we default to clang for 10.8 and later.
	// See golang.org/issue/5822.
	//
	// Roughly, OS X 10.N shows up as uname release (N+4),
	// so OS X 10.6 is uname version 10 and OS X 10.8 is uname version 12.
	if gohostos == "darwin" {
		rel := run("", CheckExit, "uname", "-r")
		if i := strings.Index(rel, "."); i >= 0 {
			rel = rel[:i]

		}
		osx, _ := strconv.Atoi(rel)
		if osx <= 6+4 {
			goextlinkenabled = "0"
		}
		if osx >= 8+4 {
			defaultclang = true
		}
	}

	if len(os.Args) > 1 && os.Args[1] == "-check-goarm" {
		useVFPv1() // might fail with SIGILL
		println("VFPv1 OK.")
		useVFPv3() // might fail with SIGILL
		println("VFPv3 OK.")
		os.Exit(0)
	}
	xinit()
	xmain()
	xexit(0)
}

// The OS-specific main calls into the portable code here.
func xmain() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd := os.Args[1]
	os.Args = os.Args[1:] // for flag parsing during cmd
	for _, ct := range cmdtab {
		if ct.name == cmd {
			flag.Usage = func() {
				fmt.Fprintf(os.Stderr, "usage: go tool dist %s [options]\n", cmd)
				flag.PrintDefaults()
				os.Exit(2)
			}
			ct.f()
			return
		}
	}

	xprintf("unknown command %s\n", cmd)
	usage()
}

func xflagparse(maxargs int) {
	flag.Var((*count)(&vflag), "v", "verbosity")
	flag.Parse()
	if maxargs >= 0 && flag.NArg() > maxargs {
		flag.Usage()
	}
}

func _p(args ...interface{}) {
	if !_p_open {
		return
	}
	fmt.Println(args...)
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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
	p("slash=" + slash)
	gohostos = runtime.GOOS
	p("gohostos=", gohostos)

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
	p("gohostarch=" + gohostarch)
	out := run("", CheckExit, "uname", "-m")
	p("out=" + out)
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
}

func p(args ...interface{}) {
	fmt.Println(args...)
}

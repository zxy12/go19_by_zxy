// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Initialization for any invocation.

// The usual variables.
var (
	goarch                 string
	gobin                  string
	gohostarch             string
	gohostos               string
	goos                   string
	goarm                  string
	go386                  string
	goroot                 string
	goroot_final           string
	goextlinkenabled       string
	gogcflags              string // For running built compiler
	workdir                string
	tooldir                string
	oldgoos                string
	oldgoarch              string
	slash                  string
	exe                    string
	defaultcc              string
	defaultcflags          string
	defaultldflags         string
	defaultcxxtarget       string
	defaultcctarget        string
	defaultpkgconfigtarget string
	rebuildall             bool
	defaultclang           bool

	vflag int // verbosity
)

// The known architectures.
var okgoarch = []string{
	"386",
	"amd64",
	"amd64p32",
	"arm",
	"arm64",
	"mips",
	"mipsle",
	"mips64",
	"mips64le",
	"ppc64",
	"ppc64le",
	"s390x",
}

// The known operating systems.
var okgoos = []string{
	"darwin",
	"dragonfly",
	"linux",
	"android",
	"solaris",
	"freebsd",
	"nacl",
	"netbsd",
	"openbsd",
	"plan9",
	"windows",
}

// find reports the first index of p in l[0:n], or else -1.
func find(p string, l []string) int {
	for i, s := range l {
		if p == s {
			return i
		}
	}
	return -1
}

// rmworkdir deletes the work directory.
func rmworkdir() {
	if vflag > 1 {
		errprintf("rm -rf %s\n", workdir)
	}
	xremoveall(workdir)
}

// xinit handles initialization of the various global state, like goroot and goarch.
func xinit() {
	goroot = os.Getenv("GOROOT")
	if slash == "/" && len(goroot) > 1 || slash == `\` && len(goroot) > 3 {
		// if not "/" or "c:\", then strip trailing path separator
		goroot = strings.TrimSuffix(goroot, slash)
	}
	_p(goroot)

	if goroot == "" {
		fatal("$GOROOT must be set")
	}

	goroot_final = os.Getenv("GOROOT_FINAL")
	if goroot_final == "" {
		goroot_final = goroot
	}

	b := os.Getenv("GOBIN")
	if b == "" {
		b = goroot + slash + "bin"
	}
	gobin = b

	b = os.Getenv("GOOS")
	if b == "" {
		b = gohostos
	}
	goos = b

	if find(goos, okgoos) < 0 {
		fatal("unknown $GOOS %s", goos)
	}

	b = os.Getenv("GOARM")
	if b == "" {
		b = xgetgoarm()
	}
	goarm = b

	b = os.Getenv("GO386")
	if b == "" {
		if cansse2() {
			b = "sse2"
		} else {
			b = "387"
		}
	}
	go386 = b
	_p("go386=", go386)

	p := pathf("%s/src/all.bash", goroot)
	_p(p)
	if !isfile(p) {
		fatal("$GOROOT is not set correctly or not exported\n"+
			"\tGOROOT=%s\n"+
			"\t%s does not exist", goroot, p)
	}

	b = os.Getenv("GOHOSTARCH")
	if b != "" {
		gohostarch = b
	}

	_p("gohostarch=", gohostarch)
	if find(gohostarch, okgoarch) < 0 {
		fatal("unknown $GOHOSTARCH %s", gohostarch)
	}

	b = os.Getenv("GOARCH")
	if b == "" {
		b = gohostarch
	}
	goarch = b
	if find(goarch, okgoarch) < 0 {
		fatal("unknown $GOARCH %s", goarch)
	}

	b = os.Getenv("GO_EXTLINK_ENABLED")
	if b != "" {
		if b != "0" && b != "1" {
			fatal("unknown $GO_EXTLINK_ENABLED %s", b)
		}
		goextlinkenabled = b
	}

	gogcflags = os.Getenv("BOOT_GO_GCFLAGS")

	b = os.Getenv("CC")
	if b == "" {
		// Use clang on OS X, because gcc is deprecated there.
		// Xcode for OS X 10.9 Mavericks will ship a fake "gcc" binary that
		// actually runs clang. We prepare different command
		// lines for the two binaries, so it matters what we call it.
		// See golang.org/issue/5822.
		if defaultclang {
			b = "clang"
		} else {
			b = "gcc"
		}
	}
	defaultcc = b

	defaultcflags = os.Getenv("CFLAGS")

	defaultldflags = os.Getenv("LDFLAGS")

	b = os.Getenv("CC_FOR_TARGET")
	if b == "" {
		b = defaultcc
	}
	defaultcctarget = b

	b = os.Getenv("CXX_FOR_TARGET")
	if b == "" {
		b = os.Getenv("CXX")
		if b == "" {
			if defaultclang {
				b = "clang++"
			} else {
				b = "g++"
			}
		}
	}

	defaultcxxtarget = b

	b = os.Getenv("PKG_CONFIG")
	if b == "" {
		b = "pkg-config"
	}
	defaultpkgconfigtarget = b

	// For tools being invoked but also for os.ExpandEnv.
	os.Setenv("GO386", go386)
	os.Setenv("GOARCH", goarch)
	os.Setenv("GOARM", goarm)
	os.Setenv("GOHOSTARCH", gohostarch)
	os.Setenv("GOHOSTOS", gohostos)
	os.Setenv("GOOS", goos)
	os.Setenv("GOROOT", goroot)
	os.Setenv("GOROOT_FINAL", goroot_final)

	// Make the environment more predictable.
	os.Setenv("LANG", "C")
	os.Setenv("LANGUAGE", "en_US.UTF8")

	workdir = xworkdir()
	_p(workdir)

	xatexit(rmworkdir)

	tooldir = pathf("%s/pkg/tool/%s_%s", goroot, gohostos, gohostarch)
	_p(tooldir)
}

/*
 * command implementations
 */

func usage() {
	xprintf("usage: go tool dist [command]\n" +
		"Commands are:\n" +
		"\n" +
		"banner         print installation banner\n" +
		"bootstrap      rebuild everything\n" +
		"clean          deletes all built files\n" +
		"env [-p]       print environment (-p: include $PATH)\n" +
		"install [dir]  install individual directory\n" +
		"list [-json]   list all supported platforms\n" +
		"test [-h]      run Go test(s)\n" +
		"version        print Go version\n" +
		"\n" +
		"All commands take -v flags to emit extra information.\n",
	)
	xexit(2)
}

// Version prints the Go version.
func cmdversion() {
	xflagparse(0)
	xprintf("version is %s\n", findgoversion())
}

// findgoversion determines the Go version to use in the version string.
func findgoversion() string {
	// The $GOROOT/VERSION file takes priority, for distributions
	// without the source repo.
	path := pathf("%s/VERSION", goroot)
	if isfile(path) {
		b := chomp(readfile(path))
		// Commands such as "dist version > VERSION" will cause
		// the shell to create an empty VERSION file and set dist's
		// stdout to its fd. dist in turn looks at VERSION and uses
		// its content if available, which is empty at this point.
		// Only use the VERSION file if it is non-empty.
		if b != "" {
			return b
		}
	}

	// The $GOROOT/VERSION.cache file is a cache to avoid invoking
	// git every time we run this command. Unlike VERSION, it gets
	// deleted by the clean command.
	path = pathf("%s/VERSION.cache", goroot)
	if isfile(path) {
		return chomp(readfile(path))
	}

	// Show a nicer error message if this isn't a Git repo.
	if !isGitRepo() {
		fatal("FAILED: not a Git repo; must put a VERSION file in $GOROOT")
	}

	// Otherwise, use Git.
	// What is the current branch?
	branch := chomp(run(goroot, CheckExit, "git", "rev-parse", "--abbrev-ref", "HEAD"))

	// What are the tags along the current branch?
	tag := "devel"
	precise := false

	// If we're on a release branch, use the closest matching tag
	// that is on the release branch (and not on the master branch).
	if strings.HasPrefix(branch, "release-branch.") {
		tag, precise = branchtag(branch)
	}

	if !precise {
		// Tag does not point at HEAD; add hash and date to version.
		tag += chomp(run(goroot, CheckExit, "git", "log", "-n", "1", "--format=format: +%h %cd", "HEAD"))
	}

	// Cache version.
	writefile(tag, path, 0)

	return tag
}

// Remove trailing spaces.
func chomp(s string) string {
	return strings.TrimRight(s, " \t\r\n")
}

// isGitRepo reports whether the working directory is inside a Git repository.
func isGitRepo() bool {
	// NB: simply checking the exit code of `git rev-parse --git-dir` would
	// suffice here, but that requires deviating from the infrastructure
	// provided by `run`.
	gitDir := chomp(run(goroot, 0, "git", "rev-parse", "--git-dir"))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(goroot, gitDir)
	}
	_p("gitDir=", gitDir)
	fi, err := os.Stat(gitDir)
	return err == nil && fi.IsDir()
}

func branchtag(branch string) (tag string, precise bool) {
	b := run(goroot, CheckExit, "git", "log", "--decorate=full", "--format=format:%d", "master.."+branch)
	tag = branch
	for _, line := range splitlines(b) {
		// Each line is either blank, or looks like
		//	  (tag: refs/tags/go1.4rc2, refs/remotes/origin/release-branch.go1.4, refs/heads/release-branch.go1.4)
		// We need to find an element starting with refs/tags/.
		i := strings.Index(line, " refs/tags/")
		if i < 0 {
			continue
		}
		i += len(" refs/tags/")
		// The tag name ends at a comma or paren (prefer the first).
		j := strings.Index(line[i:], ",")
		if j < 0 {
			j = strings.Index(line[i:], ")")
		}
		if j < 0 {
			continue // malformed line; ignore it
		}
		tag = line[i : i+j]
		if i == 0 {
			precise = true // tag denotes HEAD
		}
		break
	}
	return
}

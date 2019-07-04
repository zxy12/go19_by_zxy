// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

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

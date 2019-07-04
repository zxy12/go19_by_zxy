package main

import (
	"testing"
)

func Test_filter(t *testing.T) {
	ss := []string{"a", "b", "c", "d"}
	r := filter(ss, func(c string) bool { return c == "c" })
	if r[0] != "c" {
		t.Errorf("ret is not c %v", r)
	}
}

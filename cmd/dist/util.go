package main

import (
	_ "bytes"
	_ "fmt"
	_ "io"
	_ "io/ioutil"
	"os"
	_ "os/exec"
	_ "path/filepath"
	_ "runtime"
	_ "sort"
	_ "strconv"
	_ "strings"
	_ "sync"
	_ "time"
)

func main() {
	os.Setenv("TERM", "dumb") // disable escape codes in clang errors
}

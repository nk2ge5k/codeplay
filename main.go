package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

var tmplMap = map[string]string{
	"go": `package main

func main() {
}`,
	"php": "<?php",
}

func pref() string {
	// copied from stdlib ioutil
	r := uint32(time.Now().UnixNano() + int64(os.Getpid()))
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

func main() {
	if len(os.Args) < 2 {
		errorf("file extnension argument required")
	}
	ext := os.Args[1]
	editor := os.Getenv("EDITOR")
	if editor == "" {
		errorf("EDITOR environment variable empty")
	}

	dir := path.Join(os.TempDir(), "codeplay", pref())
	if err := os.MkdirAll(dir, 0755); err != nil {
		errorf("codeplay: failed to create directory %q: %v", dir, err)
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			errorf("codeplay: failed to remove directory %q: %v", dir, err)
		}
	}()

	fileName := "main." + ext
	filePath := path.Join(dir, fileName)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		errorf("failed to open file %q: %v", filePath, err)
	}

	if tmpl, ok := tmplMap[ext]; ok {
		// Try to copy template into file if failed ignore
		io.Copy(f, strings.NewReader(tmpl))
	}
	f.Close()

	cmd := exec.Command(editor, fileName)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		errorf("failed to run %s: %v", editor, err)
	}
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "codeplay: "+format+"\n", args...)
	os.Exit(1)
}

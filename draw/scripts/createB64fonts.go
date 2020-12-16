package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	gopath := os.Getenv("GOPATH")
	staticRoot := filepath.Join(gopath, "src", "github.com", "djangulo", "gotcha", "draw", "static")
	inconsolata := filepath.Join(staticRoot, "inconsolata")

	args := make([]interface{}, 0)
	for _, p := range []string{
		filepath.Join(inconsolata, "inconsolata_gray.png"),
		filepath.Join(inconsolata, "inconsolata_black.png"),
		filepath.Join(inconsolata, "inconsolata.png"),
	} {
		fh, err := os.Open(p)
		if err != nil {
			panic(err)
		}
		img, err := png.Decode(fh)
		if err != nil {
			panic(err)
		}
		var buf bytes.Buffer
		png.Encode(&buf, img)

		args = append(args, filepath.Base(p))
		args = append(args, base64.StdEncoding.EncodeToString(buf.Bytes()))

		fh.Close()
	}
	data := fmt.Sprintf(template, args...)
	fpath := filepath.Join(gopath, "src", "github.com", "djangulo", "gotcha", "draw", "b64assets.go")
	os.RemoveAll(fpath)
	fh, err := os.Create(fpath)
	if err != nil {
		panic(err)
	}
	_, err = fh.WriteString(data)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
}

var template = `package draw

var b64assets = map[string]string{
	%q: %q,
	%q: %q,
	%q: %q,
}

`

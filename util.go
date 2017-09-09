package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
)

const maxFileNameLen = 24

// translated from Java to Go
// https://stackoverflow.com/questions/3758606
func byteSize(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	}
	exp := math.Log(float64(bytes)) / math.Log(1024)
	pre := string("KMGTPE"[int(exp)-1]) + "i"
	return fmt.Sprintf("%.1f %sB",
		float64(bytes)/(1024*exp),
		pre)
}

// don't let file names exceed `marFileNameLen` chars
func shortenFile(name string) string {
	if len(name) <= maxFileNameLen {
		return name
	}
	ext := path.Ext(name)
	extl := len(ext) + 1 // eg.: ⋯.png
	first := name[:maxFileNameLen-extl-1]
	return first + "⋯" + ext
}

// from https://blog.sgmansfield.com/2015/12/goroutine-ids/
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

// download file if not yet exists
func dl(local, remote string) error {
	if _, err := os.Stat(local); os.IsNotExist(err) {
		f, err := os.Create(local)
		if err != nil {
			return err
		}

		resp, err := http.Get(remote)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

package scanner

import (
	"bytes"
	"compress/gzip"
	"testing"
)

func TestGzipLimit(t *testing.T) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(bytes.Repeat([]byte("x"), maxBody+1))
	w.Close()
	if _, e := decodeBody(b.Bytes(), "gzip"); e == nil {
		t.Fatal("decompression limit bypass")
	}
	if _, e := decodeBody([]byte("bad"), "gzip"); e == nil {
		t.Fatal("invalid gzip accepted")
	}
}

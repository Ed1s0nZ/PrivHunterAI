package scanner

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"strings"
)

func decodeBody(body []byte, encoding string) ([]byte, error) {
	if len(body) > maxBody {
		return nil, errors.New("body too large")
	}
	if strings.EqualFold(encoding, "gzip") {
		reader, e := gzip.NewReader(bytes.NewReader(body))
		if e != nil {
			return nil, e
		}
		defer reader.Close()
		output, e := io.ReadAll(io.LimitReader(reader, maxBody+1))
		if e != nil || len(output) > maxBody {
			return nil, errors.New("decoded body too large or invalid")
		}
		return output, nil
	}
	if encoding != "" && encoding != "identity" {
		return nil, errors.New("unsupported encoding")
	}
	return body, nil
}

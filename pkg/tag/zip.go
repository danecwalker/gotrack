package tag

import (
	"compress/gzip"
	"compress/zlib"
	"net/http"
)

func Gzip(w http.ResponseWriter, content []byte) {
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(http.StatusOK)
	gw := gzip.NewWriter(w)
	defer gw.Close()
	gw.Write(content)
}

func Deflate(w http.ResponseWriter, content []byte) {
	w.Header().Set("Content-Encoding", "deflate")
	w.WriteHeader(http.StatusOK)
	z := zlib.NewWriter(w)
	defer z.Close()
	z.Write(content)
}

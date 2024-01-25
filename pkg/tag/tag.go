package tag

import (
	"fmt"
	"net/http"
	"strings"
)

func ApplyCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func HandleTag(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path[len(r.URL.Path)-6:]
	if upath == "min.js" || upath == "all.js" {
		switch upath[:3] {
		case "min":
			MinTag(w, r)
			return
		case "all":
			AllTag(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(fmt.Sprintf("404 page not found: %s", r.URL.Path)))
}

// AllTag returns the tag with all features and only allows the debug option to be set. e.g. /tag/all.js?debug=true
func AllTag(w http.ResponseWriter, r *http.Request) {
	options := &BuildOptions{
		IsDebug:        false,
		IncludeAll:     true,
		IncludeRevenue: true,
	}

	flags := r.URL.Query()
	if len(flags) > 0 {
		for k, v := range flags {
			switch k {
			case "debug":
				options.IsDebug = v[0] == "true"
			}
		}
	}

	c, err := buildJS(options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	ApplyCors(w)
	w.Header().Set("Content-Type", "application/javascript")
	w.WriteHeader(http.StatusOK)
	w.Write(c)
}

// MinTag returns the tag with only the minimum features and allows every option to be set. e.g. /tag/min.js?debug=true, /tag/min.js?debug=true&f=links,revenue
func MinTag(w http.ResponseWriter, r *http.Request) {
	options := &BuildOptions{
		IsDebug:        false,
		IncludeAll:     false,
		IncludeRevenue: false,
	}

	flags := r.URL.Query()
	if len(flags) > 0 {
		for k, v := range flags {
			switch k {
			case "debug":
				options.IsDebug = v[0] == "true"
			}
		}
	}

	c, err := buildJS(options)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	ApplyCors(w)
	w.Header().Set("Content-Type", "application/javascript")
	accept := r.Header.Get("Accept-Encoding")
	if strings.Contains(accept, "gzip") {
		Gzip(w, c)
		return
	}
	if strings.Contains(accept, "deflate") {
		Deflate(w, c)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(c)
}

package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/danecwalker/analytics/pkg/analytics"
	"github.com/danecwalker/analytics/pkg/store"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

const port int = 3000

func main() {
	r := http.NewServeMux()

	// Create a new in-memory store.
	store := store.NewInMemoryStore()

	// Register the handler for the /track endpoint.
	r.HandleFunc("/event", analytics.TrackEvent(store))
	r.HandleFunc("/js/track.js", func(w http.ResponseWriter, r *http.Request) {
		opts := map[string]interface{}{
			"dev":        false,
			"compatible": false,
			"skip":       false,
			"manual":     false,
			"props":      false,
			"custom":     false,
			"outbound":   false,
			"downloads":  false,
			"revenue":    false,
			"name":       "track",
			"api":        "/event",
		}
		features := r.URL.Query().Get("f")
		if features != "" {
			options := strings.Split(features, ",")

			for _, o := range options {
				if o == "all" {
					opts["custom"] = true
					opts["outbound"] = true
					opts["downloads"] = true
					opts["props"] = true
					opts["revenue"] = true
				} else {
					opts[o] = true
				}
			}
		}
		var body bytes.Buffer
		t := template.Must(template.New("script.js.tmpl").Funcs(template.FuncMap{
			"orany": func(a ...bool) bool {
				for _, b := range a {
					if b {
						return true
					}
				}
				return false
			},
		}).ParseFiles("pkg/js/script.js.tmpl", "pkg/js/customEvents.js.tmpl"))
		t.Execute(&body, opts)
		m := minify.New()
		m.Add("application/javascript", &js.Minifier{
			KeepVarNames: false,
		})
		minifiedCode, err := m.String("application/javascript", body.String())
		// obf := hunterjsobfuscator.NewObfuscator(minifiedCode)
		// obfuscatedCode := obf.Obfuscate()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Write([]byte(minifiedCode))
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("page.html.tmpl").ParseFiles("cmd/page.html.tmpl"))
		w.WriteHeader(http.StatusOK)
		t.Execute(w, nil)
	})
	r.HandleFunc("/another", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		dur, _ := store.VisitDuration()
		if dur >= 60 {
			w.Write([]byte(fmt.Sprintf("%dm%ds", int(dur/60), int(dur)%60)))
		} else {
			w.Write([]byte(fmt.Sprintf("%ds", int(dur)%60)))
		}
	})

	// Start the server on port 3000.
	addr := "127.0.0.1"
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		re := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(?:\:\d+|\/\d+)?`)
		if matches := re.FindStringSubmatch(a.String()); len(matches) > 0 && (matches[1] != "127.0.0.1" || matches[1] == "localhost") {
			addr = re.FindStringSubmatch(a.String())[1]
			fmt.Println(addr)
		}
	}
	log.Printf("Starting server on http://%s:%d", addr, port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		log.Fatal(err)
	}
}

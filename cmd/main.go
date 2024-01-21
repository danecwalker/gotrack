package main

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/danecwalker/analytics/pkg/analytics"
	"github.com/danecwalker/analytics/pkg/store"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

func main() {
	r := http.NewServeMux()

	// Create a new in-memory store.
	store := store.NewInMemoryStore()

	// Register the handler for the /track endpoint.
	r.HandleFunc("/event", analytics.TrackEvent(store))
	r.HandleFunc("/js/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/js/"):]
		if strings.HasPrefix(path, "track") && strings.HasSuffix(path, ".js") {
			options := strings.Split(path, ".")
			opts := map[string]interface{}{
				"dev":        false,
				"compatible": false,
				"skip":       false,
				"manual":     false,
				"props":      false,
				"tagged":     false,
				"outbound":   false,
				"downloads":  false,
				"revenue":    false,
				"name":       "track",
				"api":        "/event",
			}
			if len(options) > 2 {
				for _, o := range options[1 : len(options)-1] {
					if o == "all" {
						opts["tagged"] = true
						opts["outbound"] = true
						opts["downloads"] = true
						opts["props"] = true
						opts["revenue"] = true
					} else if o == "revenue" {
						opts["revenue"] = true
						opts["tagged"] = true
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
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("page.html.tmpl").ParseFiles("cmd/page.html.tmpl"))
		w.WriteHeader(http.StatusOK)
		t.Execute(w, nil)
	})
	r.HandleFunc("/another", func(w http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("page.html.tmpl").ParseFiles("cmd/page.html.tmpl"))
		w.WriteHeader(http.StatusOK)
		t.Execute(w, nil)
	})

	// Start the server on port 3000.
	log.Println("Starting server on port 3000")
	http.ListenAndServe(":3000", r)
}

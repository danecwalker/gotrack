package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"regexp"

	"github.com/danecwalker/analytics/pkg/analytics"
	"github.com/danecwalker/analytics/pkg/tag"
)

const port int = 3000

func main() {
	r := http.NewServeMux()

	r.HandleFunc("/e", analytics.HandleTrackEvent)
	r.HandleFunc("/tag/", tag.HandleTag)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		t := template.Must(template.New("page.html.tmpl").ParseFiles("cmd/main/page.html.tmpl"))
		t.Execute(w, map[string]interface{}{})
	})

	r.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		t := template.Must(template.New("store.html.tmpl").ParseFiles("cmd/main/store.html.tmpl"))
		t.Execute(w, map[string]interface{}{})
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

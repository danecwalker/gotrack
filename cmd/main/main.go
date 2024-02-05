package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/danecwalker/gotrack/pkg/analytics"
	"github.com/danecwalker/gotrack/pkg/store/sqlite"
	"github.com/danecwalker/gotrack/pkg/tag"
)

const port int = 3000

func main() {
	r := http.NewServeMux()
	s, err := sqlite.NewSqlite("./cmd/main/analytics.db")
	if err != nil {
		log.Fatal(err)
	}

	// wait for ctrl+c to close db connection
	// defer func() {
	// 	_ = s.Close()
	// }()

	r.HandleFunc("/e", analytics.HandleTrackEvent(s))
	r.HandleFunc("/api/v1/stats", analytics.GetStats(s))
	r.HandleFunc("/api/v1/graph", analytics.GraphStats(s))
	r.HandleFunc("/tag/", tag.HandleTag)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Accept-CH", "Sec-CH-UA-Platform, Sec-CH-UA, Sec-CH-UA-Mobile")
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
	if os.Getenv("GO_ENV") == "dev" {
		log.Println("Running in dev mode")
	}
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		log.Fatal(err)
	}
}

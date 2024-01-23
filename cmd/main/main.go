package main

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"regexp"
	"strings"

	"github.com/danecwalker/analytics/pkg/analytics"
	"github.com/danecwalker/analytics/pkg/tag"
)

const port int = 3000

func main() {
	r := http.NewServeMux()

	r.HandleFunc("/e", analytics.HandleTrackEvent)
	r.HandleFunc("/tag.js", tag.HandleTag)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		t := template.Must(template.New("page.html.tmpl").ParseFiles("cmd/main/page.html.tmpl"))
		t.Execute(w, map[string]interface{}{})
	})

	r.HandleFunc("/mail", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")

		// resolve the ip of the recipient mail server
		mx, err := net.LookupMX(email[strings.Index(email, "@")+1:])
		if err != nil {
			log.Fatal(err)
		}

		// connect to the mail server
		doSend(email, mx[0].Host)

		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusSeeOther)
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

func doSend(email, host string) {
	// get mx record for the recipient
	mx, err := net.LookupMX(email[strings.Index(email, "@")+1:])
	if err != nil {
		log.Fatal(err)
	}
	mxHost := mx[0].Host

	for _, a := range mx {
		log.Println(a.Host)
	}

	// tls
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	// connect to the remote smtp server
	conn, err := tls.Dial("tcp", mxHost+":25", tlsconfig)
	if err != nil {
		log.Fatal(err)
	}

	t := textproto.NewConn(conn)

	c, l, _ := t.ReadCodeLine(220)
	log.Println(c, l)

}

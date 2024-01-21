package analytics

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/danecwalker/analytics/pkg/event"
	"github.com/danecwalker/analytics/pkg/store"
	"github.com/mileusna/useragent"
)

func TrackEvent(store store.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			// Return a 405 Method Not Allowed if the request method is not POST.
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
			return
		}

		// Parse the request body into an event.Event.
		var ev event.Event
		err := json.NewDecoder(r.Body).Decode(&ev)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid request body"))
			return
		}

		ua := useragent.Parse(r.Header.Get("User-Agent"))

		ip, err := getIp(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to get IP"))
			return
		}

		ev.Session = hash((time.Now().Unix() / (60 * 60 * 24)), ip, r.Header.Get("User-Agent"))
		ev.Browser = ua.Name
		ev.OS = ua.OS
		ev.Device = event.Desktop
		if ua.Mobile {
			ev.Device = event.Mobile
		} else if ua.Tablet || (ua.Desktop && ev.MaxNumTouches > 2) {
			ev.Device = event.Tablet
		}

		u, err := url.Parse(ev.Url)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to parse URL"))
			return
		}

		ref,
			utm_source,
			utm_medium,
			utm_campaign,
			utm_term,
			utm_content := u.Query().Get("ref"),
			u.Query().Get("utm_source"),
			u.Query().Get("utm_medium"),
			u.Query().Get("utm_campaign"),
			u.Query().Get("utm_term"),
			u.Query().Get("utm_content")

		if ref != "" || (utm_source != "" && (utm_medium != "" || utm_campaign != "" || utm_term != "" || utm_content != "")) {
			ev.UTM = &event.UTM{}

			if ref != "" {
				ev.UTM.Referrer = ref
			}
			if utm_source != "" {
				ev.UTM.Source = utm_source
				if utm_medium != "" {
					ev.UTM.Medium = utm_medium
				}
				if utm_campaign != "" {
					ev.UTM.Campaign = utm_campaign
				}
				if utm_term != "" {
					ev.UTM.Term = utm_term
				}
				if utm_content != "" {
					ev.UTM.Content = utm_content
				}
			}
		}

		b, _ := json.MarshalIndent(ev, "", "  ")
		fmt.Println(string(b))

		// Save the event to the store.
		err = store.Save(ev)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to save event"))
			return
		}
		// Return a 200 OK response.
		w.WriteHeader(http.StatusOK)
	}
}

func hash(i int64, remote, ua string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d%s%s", i, remote, ua))))
}

func getIp(r *http.Request) (ip string, err error) {
	// Get the IP address from the request headers. Check X-REAL-IP first, then X-FORWARDED-FOR. If neither are present, use REMOTE-ADDR. Cloudflare will set X-FORWARDED-FOR, but not X-REAL-IP. AWS will set X-REAL-IP, but not X-FORWARDED-FOR.
	if r.Header.Get("X-REAL-IP") != "" {
		ip = r.Header.Get("X-REAL-IP")
	} else if r.Header.Get("X-FORWARDED-FOR") != "" {
		ip = r.Header.Get("X-FORWARDED-FOR")
	} else {
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return
		}
	}
	return
}

package analytics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/danecwalker/gotrack/pkg/store"
	"github.com/danecwalker/gotrack/pkg/tag"
)

func GetStats(store store.DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 method not allowed"))
			return
		}

		d := r.URL.Query().Get("date")
		now := time.Now().UTC()
		if d != "" {
			pd, err := strconv.ParseInt(d, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			n := time.Unix(pd, 0)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
			now = n
		}

		duration := parsePeriod(r.URL.Query().Get("period"))
		last := now.Add(-duration)
		fmt.Println(last, now)
		stats, err := store.GetStats(last, now)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		prev_stats, err := store.GetStats(last.Add(-duration), last)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		res := stats.Calculate(prev_stats)

		tag.ApplyCors(w)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<div hx-get="/api/v1/stats?period=30d" hx-swap="outerHTML">`))
			w.Write([]byte(`<h2>` + fmt.Sprint(res.PageViews.Value) + "<span style='font-size: 1rem; margin-left: 2rem;'>" + fmt.Sprint(res.PageViews.Change) + `</span></h2>`))
			w.Write([]byte(`<h2>` + fmt.Sprint(res.Visitors.Value) + "<span style='font-size: 1rem; margin-left: 2rem;'>" + fmt.Sprint(res.Visitors.Change) + `</span></h2>`))
			w.Write([]byte(`<h2>` + fmt.Sprint(res.Bounces.Value) + "<span style='font-size: 1rem; margin-left: 2rem;'>" + fmt.Sprint(res.Bounces.Change) + `</span></h2>`))
			w.Write([]byte(`<h2>` + fmt.Sprint(res.AverageSessionLength.Value) + "<span style='font-size: 1rem; margin-left: 2rem;'>" + fmt.Sprint(res.AverageSessionLength.Change) + `</span></h2>`))
			w.Write([]byte(`</div>`))
			return
		}

		b, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

func GraphStats(store store.DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 method not allowed"))
			return
		}

		now := time.Now().UTC()

		switch r.URL.Query().Get("period") {
		case "24h":
			now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC)
		default:
			now = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		}

		duration := parsePeriod(r.URL.Query().Get("period"))
		last := now.Add(-duration)
		gr, err := store.GetViewsAndVisits(r.URL.Query().Get("period"), last, now)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		tag.ApplyCors(w)

		b, err := json.Marshal(gr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}
}

func parsePeriod(p string) time.Duration {
	switch p {
	case "hour":
		return time.Hour
	case "24h":
		return 24 * time.Hour
	case "7d":
		return 7 * 24 * time.Hour
	case "30d":
		return 30 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

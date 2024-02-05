package analytics

import (
	"net/http"

	"github.com/danecwalker/gotrack/pkg/event"
	"github.com/danecwalker/gotrack/pkg/store"
	"github.com/danecwalker/gotrack/pkg/tag"
)

func HandleTrackEvent(store store.DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("405 method not allowed"))
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 bad request"))
			return
		}

		ev := &event.Event{}
		s, we, err := ev.Parse(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if err := store.InsertSession(s); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if err := store.InsertEvent(we); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		tag.ApplyCors(w)
		w.WriteHeader(http.StatusAccepted)
	}
}

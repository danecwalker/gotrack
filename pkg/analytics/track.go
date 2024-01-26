package analytics

import (
	"fmt"
	"net/http"

	"github.com/danecwalker/analytics/pkg/event"
	"github.com/danecwalker/analytics/pkg/tag"
)

func HandleTrackEvent(w http.ResponseWriter, r *http.Request) {
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
	err := ev.Parse(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Println(ev)
	tag.ApplyCors(w)
	w.WriteHeader(http.StatusAccepted)
}

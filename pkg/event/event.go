package event

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danecwalker/analytics/pkg/utils"
)

type Event struct {
	EventName    string                 `json:"n"`
	Url          string                 `json:"u"`
	Referrer     string                 `json:"r"`
	Props        map[string]interface{} `json:"p"`
	ViewportSize string                 `json:"v"`
	Revenue      map[string]interface{} `json:"$"`
}

func (e *Event) Parse(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(e); err != nil {
		return err
	}

	s := NewSession()
	s.ParseViewportSize(e.ViewportSize)
	s.ParseLanguage(r.Header.Get("Accept-Language"))
	s.ParseUA(r.Header.Get("User-Agent"), r.Header.Get("Sec-CH-UA-Platform"), r.Header.Get("Sec-CH-UA"))

	ev := NewWEvent()
	ev.Parse(e)

	fmt.Println(utils.PrettyJson(ev))
	fmt.Println(utils.PrettyJson(s))

	return nil
}

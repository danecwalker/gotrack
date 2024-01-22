package analytics

import (
	"fmt"
	"io"
	"net/http"

	"github.com/danecwalker/analytics/pkg/tag"
	"github.com/tidwall/pretty"
)

func HandleTrackEvent(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	fmt.Println(string(pretty.Color(pretty.PrettyOptions(b, &pretty.Options{
		Width:    80,
		Prefix:   "",
		Indent:   "  ",
		SortKeys: false,
	}), pretty.TerminalStyle)))
	tag.ApplyCors(w)
	w.WriteHeader(http.StatusAccepted)
}

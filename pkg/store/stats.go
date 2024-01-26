package store

type Stats struct {
	PageViews            int `json:"page_views"`
	Visitors             int `json:"visitors"`
	Bounces              int `json:"bounces"`
	AverageSessionLength int `json:"average_session_length"`
}

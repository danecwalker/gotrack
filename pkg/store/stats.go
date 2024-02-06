package store

type Stats struct {
	PageViews            int `json:"page_views"`
	Visitors             int `json:"visitors"`
	Bounces              int `json:"bounces"`
	AverageSessionLength int `json:"average_session_length"`
}

type Diff struct {
	Value  int `json:"value"`
	Change int `json:"change"`
}

type StatsDiff struct {
	PageViews            *Diff `json:"page_views"`
	Visitors             *Diff `json:"visitors"`
	Bounces              *Diff `json:"bounces"`
	AverageSessionLength *Diff `json:"average_session_length"`
}

func (s *Stats) Calculate(prev *Stats) *StatsDiff {
	if prev == nil {
		return &StatsDiff{
			PageViews:            &Diff{Value: s.PageViews, Change: 0},
			Visitors:             &Diff{Value: s.Visitors, Change: 0},
			Bounces:              &Diff{Value: s.Bounces, Change: 0},
			AverageSessionLength: &Diff{Value: s.AverageSessionLength, Change: 0},
		}
	}

	return &StatsDiff{
		PageViews:            &Diff{Value: s.PageViews, Change: s.PageViews - prev.PageViews},
		Visitors:             &Diff{Value: s.Visitors, Change: s.Visitors - prev.Visitors},
		Bounces:              &Diff{Value: s.Bounces, Change: s.Bounces - prev.Bounces},
		AverageSessionLength: &Diff{Value: s.AverageSessionLength, Change: s.AverageSessionLength - prev.AverageSessionLength},
	}
}

type Coord struct {
	X string `json:"x"`
	Y int    `json:"y"`
}

type GraphStats struct {
	Period    string   `json:"period"`
	PageViews []*Coord `json:"page_views"`
	Visitors  []*Coord `json:"visitors"`
}

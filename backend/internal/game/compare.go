package game

import "lolidle/backend/internal/champions"

type Status string

const (
	StatusMatch   Status = "match"
	StatusPartial Status = "partial"
	StatusNoMatch Status = "nomatch"
	StatusHigher  Status = "higher" // target year > guess year
	StatusLower   Status = "lower"  // target year < guess year
)

type AttributeFeedback struct {
	Status Status `json:"status"`
}

type Feedback struct {
	Gender      AttributeFeedback `json:"gender"`
	Positions   AttributeFeedback `json:"positions"`
	Species     AttributeFeedback `json:"species"`
	Resource    AttributeFeedback `json:"resource"`
	RangeType   AttributeFeedback `json:"rangeType"`
	Regions     AttributeFeedback `json:"regions"`
	ReleaseYear AttributeFeedback `json:"releaseYear"`
}

func Compare(guess, target champions.Champion) (Feedback, bool) {
	fb := Feedback{
		Gender:      compareSingle(guess.Gender, target.Gender),
		Positions:   compareMulti(guess.Positions, target.Positions),
		Species:     compareSingle(guess.Species, target.Species),
		Resource:    compareSingle(guess.Resource, target.Resource),
		RangeType:   compareSingle(guess.RangeType, target.RangeType),
		Regions:     compareMulti(guess.Regions, target.Regions),
		ReleaseYear: compareYear(guess.ReleaseYear, target.ReleaseYear),
	}
	return fb, guess.ID == target.ID
}

func compareSingle(g, t string) AttributeFeedback {
	if g == t {
		return AttributeFeedback{Status: StatusMatch}
	}
	return AttributeFeedback{Status: StatusNoMatch}
}

func compareMulti(g, t []string) AttributeFeedback {
	gs := toSet(g)
	ts := toSet(t)

	if len(gs) == len(ts) {
		equal := true
		for k := range gs {
			if _, ok := ts[k]; !ok {
				equal = false
				break
			}
		}
		if equal {
			return AttributeFeedback{Status: StatusMatch}
		}
	}

	for k := range gs {
		if _, ok := ts[k]; ok {
			return AttributeFeedback{Status: StatusPartial}
		}
	}

	return AttributeFeedback{Status: StatusNoMatch}
}

func compareYear(g, t int) AttributeFeedback {
	switch {
	case g == t:
		return AttributeFeedback{Status: StatusMatch}
	case t > g:
		return AttributeFeedback{Status: StatusHigher}
	default:
		return AttributeFeedback{Status: StatusLower}
	}
}

func toSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}

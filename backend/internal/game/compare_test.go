package game

import (
	"testing"

	"lolidle/backend/internal/champions"
)

func ch(id string, gender string, positions []string, species, resource, rangeType string, regions []string, year int) champions.Champion {
	return champions.Champion{
		ID: id, Name: id, Gender: gender, Positions: positions, Species: species,
		Resource: resource, RangeType: rangeType, Regions: regions, ReleaseYear: year,
	}
}

func TestCompare_singleAttributes(t *testing.T) {
	target := ch("ahri", "Female", []string{"Mid"}, "Vastayan", "Mana", "Ranged", []string{"Ionia"}, 2011)

	tests := []struct {
		name     string
		guess    champions.Champion
		wantGen  Status
		wantSpec Status
		wantRes  Status
		wantRng  Status
	}{
		{
			name:    "all match",
			guess:   target,
			wantGen: StatusMatch, wantSpec: StatusMatch, wantRes: StatusMatch, wantRng: StatusMatch,
		},
		{
			name:    "all different",
			guess:   ch("garen", "Male", []string{"Top"}, "Human", "None", "Melee", []string{"Demacia"}, 2010),
			wantGen: StatusNoMatch, wantSpec: StatusNoMatch, wantRes: StatusNoMatch, wantRng: StatusNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, _ := Compare(tt.guess, target)
			if fb.Gender.Status != tt.wantGen {
				t.Errorf("Gender = %s, want %s", fb.Gender.Status, tt.wantGen)
			}
			if fb.Species.Status != tt.wantSpec {
				t.Errorf("Species = %s, want %s", fb.Species.Status, tt.wantSpec)
			}
			if fb.Resource.Status != tt.wantRes {
				t.Errorf("Resource = %s, want %s", fb.Resource.Status, tt.wantRes)
			}
			if fb.RangeType.Status != tt.wantRng {
				t.Errorf("RangeType = %s, want %s", fb.RangeType.Status, tt.wantRng)
			}
		})
	}
}

func TestCompare_multiAttributes(t *testing.T) {
	target := ch("yasuo", "Male", []string{"Mid", "Top"}, "Human", "Flow", "Melee", []string{"Ionia"}, 2013)

	tests := []struct {
		name    string
		guess   champions.Champion
		wantPos Status
		wantReg Status
	}{
		{
			name:    "exact match positions and regions",
			guess:   ch("yone", "Male", []string{"Mid", "Top"}, "Spirit", "Flow", "Melee", []string{"Ionia"}, 2020),
			wantPos: StatusMatch, wantReg: StatusMatch,
		},
		{
			name:    "partial positions, exact region",
			guess:   ch("akali", "Female", []string{"Mid"}, "Human", "Energy", "Melee", []string{"Ionia"}, 2010),
			wantPos: StatusPartial, wantReg: StatusMatch,
		},
		{
			name:    "no match positions, no match region",
			guess:   ch("garen", "Male", []string{"Jungle"}, "Human", "None", "Melee", []string{"Demacia"}, 2010),
			wantPos: StatusNoMatch, wantReg: StatusNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, _ := Compare(tt.guess, target)
			if fb.Positions.Status != tt.wantPos {
				t.Errorf("Positions = %s, want %s", fb.Positions.Status, tt.wantPos)
			}
			if fb.Regions.Status != tt.wantReg {
				t.Errorf("Regions = %s, want %s", fb.Regions.Status, tt.wantReg)
			}
		})
	}
}

func TestCompare_releaseYear(t *testing.T) {
	target := ch("ahri", "Female", []string{"Mid"}, "Vastayan", "Mana", "Ranged", []string{"Ionia"}, 2011)

	tests := []struct {
		name      string
		guessYear int
		want      Status
	}{
		{"same year", 2011, StatusMatch},
		{"target older than guess", 2020, StatusLower},
		{"target newer than guess", 2009, StatusHigher},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guess := target
			guess.ReleaseYear = tt.guessYear
			fb, _ := Compare(guess, target)
			if fb.ReleaseYear.Status != tt.want {
				t.Errorf("ReleaseYear = %s, want %s", fb.ReleaseYear.Status, tt.want)
			}
		})
	}
}

func TestCompare_correctFlag(t *testing.T) {
	target := ch("ahri", "Female", []string{"Mid"}, "Vastayan", "Mana", "Ranged", []string{"Ionia"}, 2011)

	_, correct := Compare(target, target)
	if !correct {
		t.Error("expected correct=true when guessing the target")
	}

	other := ch("yasuo", "Male", []string{"Mid"}, "Human", "Flow", "Melee", []string{"Ionia"}, 2013)
	_, correct = Compare(other, target)
	if correct {
		t.Error("expected correct=false for different champion")
	}
}

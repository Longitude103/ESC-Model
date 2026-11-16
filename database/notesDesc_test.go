package database

import "testing"

func TestDescriptionFromNotes(t *testing.T) {
	run := []ResultsNote{{1, "Desc: 2024 base run"}, {2, "Start Year: 1997"}, {3, "End Year: 2024"}}
	if d, err := descriptionFromNotes(run); err != nil || d != "Desc: 2024 base run" {
		t.Errorf("run only: got %q, %v", d, err)
	}

	scenario := append(run, ResultsNote{4, "Robust Review Years: 2018, 2020"}, ResultsNote{5, "Robust Review Description: task3 npnrd"})
	if d, err := descriptionFromNotes(scenario); err != nil || d != "task3 npnrd" {
		t.Errorf("scenario: got %q, %v", d, err)
	}

	blankScenario := append(run, ResultsNote{4, "Robust Review Description: "})
	if d, _ := descriptionFromNotes(blankScenario); d != "Desc: 2024 base run" {
		t.Errorf("blank scenario description should fall back, got %q", d)
	}

	if _, err := descriptionFromNotes([]ResultsNote{{1, "SS"}, {2, "grid=2"}}); err == nil {
		t.Error("expected an error with no description (and no panic on short notes)")
	}
}

func TestDryCellFileType(t *testing.T) {
	cases := map[string]int{"np": 128, "SP": 129, "": 101, "unw": 101}
	for nrd, want := range cases {
		if got := (DryCell{Nrd: nrd}).FileType(); got != want {
			t.Errorf("DryCell{Nrd: %q}.FileType() = %d, want %d", nrd, got, want)
		}
	}
}

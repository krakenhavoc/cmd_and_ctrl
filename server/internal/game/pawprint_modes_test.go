package game

import (
	"errors"
	"testing"
)

// pawprint_modes_test.go — CR 700.2i, #4786: "its controller can choose
// any number of modes such that the total number of pawprint symbols
// listed for the chosen modes is not greater than the specified number."
// validateModes is the announce gate; it bounds a pawprint spell by the
// {P} its chosen modes spend, a repeat counted again (CR 700.2d).

func pawprintSpec() *ModeSpec {
	return &ModeSpec{
		Prompt: "Choose up to five {P} worth of modes", Min: 0, Max: 5, Repeatable: true, Pawprints: 5,
		Options: []ModeOption{{Label: "A.", Pawprints: 1}, {Label: "B.", Pawprints: 2}, {Label: "C.", Pawprints: 3}},
	}
}

func TestPawprintBudgetBoundsTheSelectionByItsPawprints(t *testing.T) {
	ms := pawprintSpec()
	for _, tc := range []struct {
		modes []int
		ok    bool
	}{
		{nil, true},                      // up to five: none at all
		{[]int{2, 1}, true},              // 3 + 2
		{[]int{0, 0, 0, 0, 0}, true},     // five times the {P} mode
		{[]int{1, 1, 0}, true},           // 2 + 2 + 1
		{[]int{2, 2}, false},             // 6: two bullets, under Max, over the budget
		{[]int{2, 1, 0}, false},          // 6
		{[]int{1, 1, 1}, false},          // 6
		{[]int{0, 0, 0, 0, 0, 0}, false}, // 6
	} {
		err := validateModes(ms, ms.Min, ms.Max, tc.modes)
		if (err == nil) != tc.ok {
			t.Errorf("modes %v (%d {P}): err = %v, want ok=%v", tc.modes, PawprintsOf(ms, tc.modes), err, tc.ok)
		}
		if err != nil && !errors.Is(err, ErrInvalidParam) {
			t.Errorf("modes %v: err = %v, want ErrInvalidParam", tc.modes, err)
		}
		if ms.WithinPawprints(tc.modes) != tc.ok {
			t.Errorf("WithinPawprints(%v) = %v, want %v", tc.modes, !tc.ok, tc.ok)
		}
	}
}

func TestAModalSpecWithoutPawprintsIsBoundedByItsCountAlone(t *testing.T) {
	ms := &ModeSpec{Prompt: "Choose three", Min: 3, Max: 3, Repeatable: true,
		Options: []ModeOption{{Label: "A."}, {Label: "B."}}}
	if err := validateModes(ms, 3, 3, []int{1, 1, 1}); err != nil {
		t.Errorf("a Confluence-shaped selection was refused: %v", err)
	}
	if !ms.WithinPawprints([]int{1, 1, 1}) || PawprintsOf(ms, []int{1, 1, 1}) != 0 {
		t.Error("a spec without a budget should spend no {P} and always fit")
	}
}

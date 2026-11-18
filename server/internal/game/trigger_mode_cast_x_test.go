package game

import "testing"

// #5898: a XFromCast modal trigger's chosen-mode steps are counted and
// bounded by the source's cast X, on copies — the catalog's clause and
// its total rule keep their declared values.
func TestBindTriggerModeStepsCastXCountsAndBoundsOnCopies(t *testing.T) {
	total := &TargetTotal{Label: "have total mana value X or less", Stat: TargetTotalManaValue, MaxIsX: true}
	clause := &TargetSpec{Zones: []ZoneKind{ZoneGraveyard}, CountFromX: true, UpToX: true, Total: total}
	ms := &ModeSpec{Min: 1, Max: 1, Options: []ModeOption{{Label: "sweep"}, {Label: "return", Targets: clause}}}
	row := TriggeredAbility{XFromCast: true, Modes: ms}
	src := Card{Provenance: CastProvenance{X: 3}}

	steps := bindTriggerModeStepsCastX(row, &src, AnnouncedClauses(nil, ms, []int{1}))
	if len(steps) != 1 {
		t.Fatalf("one step, got %d", len(steps))
	}
	c := steps[0].Clause
	if c.Min != 0 || c.Max != 3 || c.CountFromX {
		t.Errorf("up to X = 3: %d..%d CountFromX=%v", c.Min, c.Max, c.CountFromX)
	}
	if c.Total == nil || c.Total.Max != 3 || c.Total.Label != "have total mana value 3 or less" {
		t.Errorf("total bound to 3: %+v", c.Total)
	}
	if total.Max != 0 || !clause.CountFromX || clause.Max != 0 {
		t.Errorf("the catalog was written: %+v %+v", total, clause)
	}

	// X = 0 (not cast): the step takes no target and is dropped.
	if got := bindTriggerModeStepsCastX(row, &Card{}, AnnouncedClauses(nil, ms, []int{1})); len(got) != 0 {
		t.Errorf("X = 0 keeps %d steps", len(got))
	}
	// A row without XFromCast is left alone.
	plain := row
	plain.XFromCast = false
	if got := bindTriggerModeStepsCastX(plain, &src, AnnouncedClauses(nil, ms, []int{1})); len(got) != 1 || !got[0].Clause.CountFromX {
		t.Errorf("a plain row was rewritten: %+v", got)
	}
}

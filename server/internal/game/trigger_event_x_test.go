package game

import "testing"

// trigger_event_x_test.go — #6085: a trigger's CountFromX clause sized by
// the triggering event's Amount (TriggeredAbility.XFromEventAmount,
// Rampaging War Mammoth's "When you cycle this card, destroy up to X
// target artifacts").

func eventXRow(upTo, flag bool) TriggeredAbility {
	spec := &TargetSpec{Mode: "permanent", Label: "up to X target artifacts", Zones: []ZoneKind{ZoneBattlefield}, CountFromX: true, UpToX: upTo}
	return TriggeredAbility{Key: "event-x probe", Watches: []EventKind{EventCycle}, Targets: spec, XFromEventAmount: flag}
}

func TestEventAmountSizesAnUpToXTriggerClause(t *testing.T) {
	g := newActiveGame(t)
	row := eventXRow(true, true)
	g.mu.Lock()
	defer g.mu.Unlock()
	got := g.triggerTargetsLocked(row, TriggerContext{Event: Event{Kind: EventCycle, Amount: 3}}, &Card{})
	if got == nil || got.Min != 0 || got.Max != 3 || got.CountFromX {
		t.Fatalf("X=3 up to X: clause %+v, want 0..3 with CountFromX cleared", got)
	}
	if !row.Targets.CountFromX || row.Targets.Max != 0 {
		t.Error("the catalog's clause was rewritten in place")
	}
	exact := g.triggerTargetsLocked(eventXRow(false, true), TriggerContext{Event: Event{Kind: EventCycle, Amount: 2}}, &Card{})
	if exact == nil || exact.Min != 2 || exact.Max != 2 {
		t.Fatalf("X=2 exactly X: clause %+v, want 2..2", exact)
	}
}

// X = 0 is no targets at all — never the unbounded Max 0 (#259).
func TestEventAmountOfZeroHasNoTargets(t *testing.T) {
	g := newActiveGame(t)
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, amount := range []int{0, -1} {
		if got := g.triggerTargetsLocked(eventXRow(true, true), TriggerContext{Event: Event{Kind: EventCycle, Amount: amount}}, &Card{}); got != nil {
			t.Errorf("X=%d: clause %+v, want none", amount, got)
		}
	}
}

// Without the flag the row's clause is returned as declared.
func TestEventAmountIsReadOnlyForAnXFromEventAmountRow(t *testing.T) {
	g := newActiveGame(t)
	row := eventXRow(true, false)
	g.mu.Lock()
	defer g.mu.Unlock()
	if got := g.triggerTargetsLocked(row, TriggerContext{Event: Event{Kind: EventCycle, Amount: 3}}, &Card{}); got != row.Targets {
		t.Errorf("clause %+v, want the row's own", got)
	}
}

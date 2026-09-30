package game

import (
	"testing"

	"github.com/google/uuid"
)

// forced_mode_count_test.go — #1655, ADR 0065's 2026-09-28 amendment,
// at the engine level: a conditional mode count that raises the
// MINIMUM with the maximum ("choose both instead", no "may"), and a
// condition that reads the optional costs announced with the choice.
// The cards (Inscription of Ruin, Depth Defiler, Prophetic Titan) are
// internal/cards/effects/optional_mode_count_test.go.

// announcedAnything is a stand-in announcement condition: it holds
// when any optional cost was announced with the choice.
var announcedAnything = ModeConditionOnAnnouncement("test/announced-anything", func(_ *Game, q ModeCountQuery) bool {
	return len(q.OptionalCosts) > 0
})

// modeHi is the upper bound modeBoundsLocked quotes a chooser with
// nothing announced — the #1590 tests' question.
func modeHi(g *Game, ms *ModeSpec, chooser uuid.UUID) int {
	_, hi := g.modeBoundsLocked(ms, ModeCountQuery{Chooser: chooser})
	return hi
}

func TestForcedModeCountRaisesTheMinimumToo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ms := (&ModeSpec{Options: make([]ModeOption, 2), Min: 1, Max: 1}).InsteadIf(2, controlsAnArtifact)

	lo, hi := g.modeBoundsLocked(ms, ModeCountQuery{Chooser: me.ID})
	if lo != 1 || hi != 1 {
		t.Fatalf("condition false: bounds %d..%d, want the printed 1..1", lo, hi)
	}
	if err := validateModes(ms, lo, hi, []int{0}); err != nil {
		t.Errorf("one bullet under the printed count: %v", err)
	}
	g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	lo, hi = g.modeBoundsLocked(ms, ModeCountQuery{Chooser: me.ID})
	if lo != 2 || hi != 2 {
		t.Fatalf("condition true: bounds %d..%d, want the forced 2..2", lo, hi)
	}
	if err := validateModes(ms, lo, hi, []int{0}); err != ErrInvalidParam {
		t.Errorf("one bullet under a forced two: %v, want ErrInvalidParam", err)
	}
	if err := validateModes(ms, lo, hi, []int{1, 0}); err != nil {
		t.Errorf("both under a forced two: %v", err)
	}
}

func TestAnyNumberIfKeepsTheMinimum(t *testing.T) {
	g := newActiveGame(t)
	ms := (&ModeSpec{Options: make([]ModeOption, 3), Min: 1, Max: 1}).AnyNumberIf(announcedAnything)
	lo, hi := g.modeBoundsLocked(ms, ModeCountQuery{OptionalCosts: []int{0}})
	if lo != 1 || hi != 3 {
		t.Fatalf("announced: bounds %d..%d, want 1..3", lo, hi)
	}
	if err := validateModes(ms, lo, hi, nil); err != ErrInvalidParam {
		t.Errorf("no bullet at all: %v, want ErrInvalidParam (the minimum is still one)", err)
	}
	if lo, hi := g.modeBoundsLocked(ms, ModeCountQuery{}); lo != 1 || hi != 1 {
		t.Errorf("nothing announced: bounds %d..%d, want 1..1", lo, hi)
	}
}

// A forced count on a TRIGGER: the prompt carries the raised minimum,
// so one bullet is refused at the answer gate (CR 603.3c).
func TestModalTriggerPromptCarriesTheForcedMinimum(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	ability := threeBulletTrigger(&ran)
	ability.Modes.InsteadIf(2, controlsAnArtifact)
	src := modalTriggerSource(t, g, me, "test-forced-modal-trigger", ability)
	g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})

	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	c := openModePick(t, g, me.ID)
	if c == nil || c.ModeMin != 2 || c.ModeMax != 2 {
		t.Fatalf("with an artifact: the prompt must force two: %+v", c)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != ErrInvalidParam {
		t.Errorf("one bullet under a forced two: %v, want ErrInvalidParam", err)
	}
	for _, sel := range ModePickSelections(c, 50) {
		if len(sel) != 2 {
			t.Errorf("the enumerator offered %v under a forced two", sel)
		}
	}
}

// CR 603.3c: a bullet with no legal target "can't be chosen", which
// is not the same as the ability being removed — a forced "choose
// both" whose second bullet cannot be filled asks for the one that
// can.
func TestForcedTriggerCountAsksOnlyForTheFillableBullets(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	ability := threeBulletTrigger(&ran)
	ability.Modes.InsteadIf(3, controlsAnArtifact)
	const oracle = "test-forced-modal-trigger-unfillable"
	id := uuid.New()
	// An ENCHANTMENT source and no creature anywhere, so the "destroy
	// target creature" bullet has no legal target.
	g.Battlefield.PushTop(Card{InstanceID: id, Name: "Modal Source", OracleID: oracle, TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	withCatalogTriggers(t, func(o string) []TriggeredAbility {
		if o != oracle {
			return nil
		}
		return []TriggeredAbility{ability}
	})

	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: id, Actor: me.ID}) })
	c := openModePick(t, g, me.ID)
	if c == nil {
		t.Fatal("the trigger was removed; CR 603.3c keeps it with the bullets that can be chosen")
	}
	if len(c.ModeOptionIndex) != 2 || c.ModeMin != 2 || c.ModeMax != 3 {
		t.Fatalf("offered %v, bounds %d..%d; want the two fillable bullets, both forced", c.ModeOptionIndex, c.ModeMin, c.ModeMax)
	}
}

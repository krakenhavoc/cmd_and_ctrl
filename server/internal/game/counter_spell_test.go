package game

import (
	"testing"

	"github.com/google/uuid"
)

// counter_spell_test.go — #4056: the CR 614 window on a spell about to
// be countered (RepEventCounterSpell). Stubbed replacements, so this
// pins the engine contract; Guile, the printed card, is pinned end to end
// in cards/effects/guile_test.go.

// counteringFor makes `p` the controller of the spell or ability that is
// resolving, which is who "a spell or ability you control would counter
// a spell" reads.
func counteringFor(g *Game, p *Player) {
	g.WithWriteLock(func() {
		g.resolving = &resolvingItem{item: &StackItem{
			ID: uuid.New(), Kind: StackItemSpell, Controller: p.ID, SourceCardID: uuid.New(),
		}}
	})
}

// counterWatcher is a test replacement of a counter: `apply` is its
// Replace. It records what it saw in `seen`.
func counterWatcher(owner uuid.UUID, label string, seen *[]ReplacementEvent, apply func(ev *ReplacementEvent, g *Game) error) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventCounterSpell},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.CounterSpellStillCounters()
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			if seen != nil {
				*seen = append(*seen, *ev)
			}
			return apply(ev, g)
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

func cancelCounter(ev *ReplacementEvent, _ *Game) error { ev.Cancel(); return nil }

func exileInstead(by uuid.UUID) func(ev *ReplacementEvent, g *Game) error {
	return func(ev *ReplacementEvent, g *Game) error {
		return g.ExileCounteredSpellInsteadForEffect(ev, by, uuid.New(), "Test Guile")
	}
}

// The window opens for a counter a resolving spell or ability makes, and
// names who is countering. Cancelled, the spell is not countered: it
// stays on the stack, no counter event is emitted, and the caller's
// continuation hears that nothing landed.
func TestACounterOpensTheWindowAndACancelLeavesTheSpellOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushSpellFor(t, g, "Instant", me, me)
	var seen []ReplacementEvent
	g.WithWriteLock(func() { g.RegisterReplacementForTest(counterWatcher(opp.ID, "cancel", &seen, cancelCounter)) })
	counteringFor(g, opp)

	var placed, told bool
	g.WithWriteLock(func() {
		err := g.CounterSpellToLibraryThenForEffect(spell, TuckOptions{ToBottom: true}, func(_ *Game, p bool) error {
			placed, told = p, true
			return nil
		})
		if err != nil {
			t.Fatalf("counter: %v", err)
		}
	})

	if len(seen) != 1 || seen[0].CounterSpellID != spell || seen[0].CounterBy != opp.ID || seen[0].Actor != opp.ID {
		t.Fatalf("the replacement saw %+v, want one counter of the spell by the opponent", seen)
	}
	if !g.Stack.Contains(spell) {
		t.Error("a cancelled counter took the spell off the stack")
	}
	if hasEventFor(g, EventCounterSpell, spell) {
		t.Error("a cancelled counter emitted EventCounterSpell")
	}
	if !told || placed {
		t.Errorf("continuation told=%v placed=%v, want told and not placed", told, placed)
	}
}

// With nothing resolving, no spell or ability is countering (the
// sandbox's counter action), so no window opens and the spell goes to
// the graveyard as it always did.
func TestACounterWithNothingResolvingOpensNoWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := pushSpellFor(t, g, "Instant", me, me)
	var seen []ReplacementEvent
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(counterWatcher(me.ID, "cancel", &seen, cancelCounter))
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if len(seen) != 0 {
		t.Errorf("the window opened with nothing resolving: %+v", seen)
	}
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("the spell was not countered into its owner's graveyard")
	}
	if !hasEventFor(g, EventCounterSpell, spell) {
		t.Error("no EventCounterSpell for a plain counter")
	}
}

// A spell that can't be countered is not countered, so there is nothing
// to replace: the window never opens for it (CR 101.2).
func TestACantBeCounteredSpellNeverOpensTheWindow(t *testing.T) {
	withCounterShields(t, CounterShieldStatic{Label: "Spells you control can't be countered."})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushShieldSource(g, me)
	spell := pushSpellFor(t, g, "Instant", me, me)
	var seen []ReplacementEvent
	g.WithWriteLock(func() { g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", &seen, exileInstead(opp.ID))) })
	counteringFor(g, opp)

	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if len(seen) != 0 {
		t.Errorf("the window opened for a spell that can't be countered: %+v", seen)
	}
	if !g.Stack.Contains(spell) {
		t.Error("an uncounterable spell left the stack")
	}
}

// A replaced counter exiles the spell WITHOUT countering it: it lands in
// exile, nothing keyed to "countered" fires, the countering player (not
// the spell's owner) holds a free cast of the exiled card, and the
// destination counter's continuation hears the spell did not reach a
// library. Every effect verb passes through the same window.
func TestAReplacedCounterExilesTheSpellUncounteredAndGrantsTheFreeCast(t *testing.T) {
	verbs := []struct {
		name    string
		counter func(g *Game, id uuid.UUID, told *bool) error
	}{
		{"counter", func(g *Game, id uuid.UUID, _ *bool) error { return g.CounterTargetForEffect(id) }},
		{"counter to its owner's hand", func(g *Game, id uuid.UUID, _ *bool) error {
			return g.CounterTargetToZoneForEffect(id, ZoneRef{Kind: ZoneHand, Owner: g.Seats[0].ID})
		}},
		{"counter to the top of the library", func(g *Game, id uuid.UUID, told *bool) error {
			return g.CounterSpellToLibraryThenForEffect(id, TuckOptions{}, func(_ *Game, placed bool) error {
				if placed {
					t.Error("a replaced counter reports the spell as placed in the library")
				}
				*told = true
				return nil
			})
		}},
	}
	for _, v := range verbs {
		t.Run(v.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			spell := pushSpellFor(t, g, "Creature — Bear", me, me)
			g.WithWriteLock(func() { g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", nil, exileInstead(opp.ID))) })
			counteringFor(g, opp)

			var told bool
			g.WithWriteLock(func() {
				if err := v.counter(g, spell, &told); err != nil {
					t.Fatalf("counter: %v", err)
				}
			})

			if g.Stack.Contains(spell) || !g.Exile.Contains(spell) {
				t.Fatal("the spell is not in exile")
			}
			if hasEventFor(g, EventCounterSpell, spell) {
				t.Error("an exiled-instead spell was reported as countered")
			}
			if _, still := g.StackMeta[spell]; still {
				t.Error("the spell's stack record outlived it")
			}
			var perm *CastPermission
			g.WithWriteLock(func() { perm = g.CastPermissionOnCardByIDForEffect(spell) })
			if perm == nil {
				t.Fatal("no free cast over the exiled card")
			}
			if perm.Player != opp.ID || perm.Cost != "{0}" || perm.Zone != ZoneExile || !perm.CastOnly {
				t.Errorf("grant = %+v, want a free cast-only permission in exile for the countering player", perm)
			}
			if v.name == "counter to the top of the library" && !told {
				t.Error("the destination counter's continuation was never told")
			}
		})
	}
}

// Two replacements that disagree are a CR 616.1 ordering question put to
// the countered spell's controller, the affected object's controller,
// with nothing moved while it is open; the order picked decides the
// outcome. Cancel first: the spell stays. Exile first: it is exiled, and
// the cancel no longer has a counter to cancel.
func TestTwoCounterReplacementsAreOrderedByTheSpellsController(t *testing.T) {
	for _, tc := range []struct {
		name        string
		exileFirst  bool
		wantInExile bool
	}{
		{"exile first", true, true},
		{"cancel first", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			spell := pushSpellFor(t, g, "Instant", me, me)
			g.WithWriteLock(func() {
				g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", nil, exileInstead(opp.ID)))
				g.RegisterReplacementForTest(counterWatcher(opp.ID, "cancel", nil, cancelCounter))
			})
			counteringFor(g, opp)

			told := 0
			g.WithWriteLock(func() {
				err := g.CounterSpellToLibraryThenForEffect(spell, TuckOptions{}, func(*Game, bool) error {
					told++
					return nil
				})
				if err != nil {
					t.Fatalf("counter: %v", err)
				}
			})

			p := onlyPrompt(t, g)
			if p.Kind != PendingChoiceReplacementOrder {
				t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceReplacementOrder)
			}
			if p.Chooser != me.ID {
				t.Errorf("chooser = %s, want the countered spell's controller %s; got %s (the controller of the replacements)",
					p.Chooser, me.ID, whichSeat(g, p.Chooser))
			}
			if !g.Stack.Contains(spell) || told != 0 {
				t.Fatalf("nothing may move or be told while the prompt is open (on stack %v, told %d)", g.Stack.Contains(spell), told)
			}

			ids := append([]ReplacementEffectID(nil), p.ReplacementEffectIDs...)
			if len(ids) != 2 {
				t.Fatalf("prompt offers %d effects, want 2", len(ids))
			}
			// The gather order is exile, then cancel.
			ordered := ids
			if !tc.exileFirst {
				ordered = []ReplacementEffectID{ids[1], ids[0]}
			}
			if err := g.ResolveReplacementOrder(p.ID, me.ID, ordered); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}

			if g.Exile.Contains(spell) != tc.wantInExile {
				t.Errorf("in exile = %v, want %v", g.Exile.Contains(spell), tc.wantInExile)
			}
			if g.Stack.Contains(spell) == tc.wantInExile {
				t.Errorf("on the stack = %v, want %v", g.Stack.Contains(spell), !tc.wantInExile)
			}
			if hasEventFor(g, EventCounterSpell, spell) {
				t.Error("a replaced counter was reported as a counter")
			}
			if told != 1 {
				t.Errorf("the continuation ran %d times, want exactly once", told)
			}
		})
	}
}

// A prompt taken away (its chooser concedes) is a terminal outcome: the
// spell is not countered, and the effect behind it is still told.
func TestAnAbandonedCounterOrderingPromptStillTellsItsContinuation(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	spell := pushSpellFor(t, g, "Instant", me, me)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(counterWatcher(opp.ID, "exile", nil, exileInstead(opp.ID)))
		g.RegisterReplacementForTest(counterWatcher(opp.ID, "cancel", nil, cancelCounter))
	})
	counteringFor(g, opp)

	told := 0
	g.WithWriteLock(func() {
		if err := g.CounterSpellToLibraryThenForEffect(spell, TuckOptions{}, func(*Game, bool) error {
			told++
			return nil
		}); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if p := onlyPrompt(t, g); p.Chooser != me.ID {
		t.Fatalf("chooser = %s, want the spell's controller", whichSeat(g, p.Chooser))
	}
	if told != 0 {
		t.Fatal("told while the prompt was open")
	}

	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if told != 1 {
		t.Errorf("the continuation ran %d times after the prompt was dropped, want 1", told)
	}
}

// A second answer on the same event changes nothing: the first
// replacement's answer stands (CR 616.1f), and a plain event refuses an
// answer meant for a counter.
func TestExileInsteadStandsOnceAndOnlyForACounter(t *testing.T) {
	g := newActiveGame(t)
	ev := &ReplacementEvent{Kind: RepEventCounterSpell}
	first, second := uuid.New(), uuid.New()
	if err := g.ExileCounteredSpellInsteadForEffect(ev, first, uuid.Nil, "one"); err != nil {
		t.Fatal(err)
	}
	if err := g.ExileCounteredSpellInsteadForEffect(ev, second, uuid.Nil, "two"); err != nil {
		t.Fatal(err)
	}
	if ev.counterInstead.mayPlayFreeBy != first {
		t.Error("the second answer replaced the first")
	}
	if ev.CounterSpellStillCounters() {
		t.Error("a replaced counter still reads as a counter")
	}
	if err := g.ExileCounteredSpellInsteadForEffect(&ReplacementEvent{Kind: RepEventMove}, first, uuid.Nil, "x"); err == nil {
		t.Error("a move accepted a counter's answer")
	}
}

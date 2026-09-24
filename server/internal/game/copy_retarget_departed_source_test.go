package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// copy_retarget_departed_source_test.go covers #1449 (ADR 0072
// amendment 2026-09-24 "#1449"): when an ABILITY is copied and its
// controller may choose new targets for the copy (CR 707.10c), the new
// targets are judged against the ability's source as it LAST EXISTED
// on the battlefield (CR 608.2h) once that source has left — at the
// offer, at the #809 refresh of the open prompt, and at the answer.
// Before, all three read the source's graveyard card, a new object
// with printed characteristics (CR 400.7).
//
// The scenarios reuse #1429's (departed_ability_source_target_test.go)
// layer 5 painter and bleacher, which apply on the battlefield only, so
// the source's last-known colour and its graveyard colour DIFFER, and
// each setup asserts that difference.

// copyRetargetScene is one ability on the stack, aimed at a plain
// creature, with a creature that has protection from red beside it.
type copyRetargetScene struct {
	g        *Game
	me       *Player
	src      uuid.UUID
	item     *StackItem
	resolved *bool
	plain    uuid.UUID
	proRed   uuid.UUID
}

// newCopyRetargetScene builds the scene. `redOnBattlefield` picks the
// source: a printed-colourless one painted red (red on the battlefield,
// colourless in the graveyard) or a printed-red one bleached
// colourless (colourless on the battlefield, red in the graveyard).
func newCopyRetargetScene(t *testing.T, kind StackItemKind, redOnBattlefield bool) *copyRetargetScene {
	t.Helper()
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	painted, bleached := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	withLKIColourPainters(t, painted, bleached)

	s := &copyRetargetScene{g: g, me: me}
	s.plain = pushColouredCreature(g, opp, "Plain", []string{"W"})
	s.proRed = pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	if redOnBattlefield {
		s.src = pushLKISource(g, me.ID, nil)
		painted[s.src] = true
		pushLKIPainter(g, me.ID, lkiPainterOracle)
	} else {
		s.src = pushLKISource(g, me.ID, []string{"R"})
		bleached[s.src] = true
		pushLKIPainter(g, me.ID, lkiBleacherOracle)
	}
	want := []string(nil)
	if redOnBattlefield {
		want = []string{"R"}
	}
	if got := liveColours(t, g, s.src); !slices.Equal(got, want) {
		t.Fatalf("setup: the source is %v on the battlefield, want %v", got, want)
	}
	s.item, s.resolved = pushTargetedAbilityItem(t, g, kind, s.src, me.ID, s.plain)
	return s
}

// kill destroys the source and checks its graveyard card has the other
// colour, without which the scenario proves nothing.
func (s *copyRetargetScene) kill(t *testing.T, redOnBattlefield bool) {
	t.Helper()
	destroy(t, s.g, s.src)
	got := liveColours(t, s.g, s.src)
	if redOnBattlefield && len(got) != 0 {
		t.Fatalf("setup: the graveyard card is %v; it must be colourless", got)
	}
	if !redOnBattlefield && !slices.Equal(got, []string{"R"}) {
		t.Fatalf("setup: the graveyard card is %v; it must be red", got)
	}
}

// copyWithNewTargets copies the item with the CR 707.10c offer and
// returns the prompt.
func (s *copyRetargetScene) copyWithNewTargets(t *testing.T) *PendingChoice {
	t.Helper()
	s.g.WithWriteLock(func() {
		if err := s.g.CopyAbilityForEffect(s.item.ID, s.me.ID, true); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})
	choice := pendingOfKind(t, s.g, PendingChoicePickTarget)
	if !slices.Contains(choice.PickTargetCards, s.plain) {
		t.Fatalf("setup: the plain creature must be offered either way: %v", choice.PickTargetCards)
	}
	return choice
}

// refresh runs the state checks, whose first act is the #809 refresh of
// every open pick_target prompt (refreshTargetChoicesLocked).
func (s *copyRetargetScene) refresh() {
	s.g.WithWriteLock(func() { s.g.runStateChecksLocked() })
}

// A source painted red dies (colourless in the graveyard). It was red
// as it last existed, so the copy may NOT be pointed at the pro-red
// creature: the prompt does not offer it, the refresh does not add it,
// and an answer naming it is refused.
func TestAbilityCopyOfADepartedRedSourceCannotRetargetPastProtectionFromRed(t *testing.T) {
	for _, kind := range []StackItemKind{StackItemTriggered, StackItemActivated} {
		t.Run(string(kind), func(t *testing.T) {
			s := newCopyRetargetScene(t, kind, true)
			s.kill(t, true)
			choice := s.copyWithNewTargets(t)

			if slices.Contains(choice.PickTargetCards, s.proRed) {
				t.Error("the copy's re-target prompt offers the pro-red creature: the source was " +
					"RED as it last existed (CR 707.10c / 608.2h) — it read the colourless graveyard card (#1449)")
			}
			s.refresh()
			if slices.Contains(choice.PickTargetCards, s.proRed) {
				t.Error("the #809 refresh widened the prompt to the pro-red creature: it read the " +
					"colourless graveyard card (#1449)")
			}
			if err := s.g.ResolvePickTargets(choice.ID, s.me.ID,
				[]TargetRef{{Kind: TargetCard, ID: s.proRed}}); err == nil {
				t.Error("the answer accepted the pro-red creature as the copy's new target: the " +
					"source was RED as it last existed (#1449)")
			}
		})
	}
}

// The reverse: a printed-red source bleached colourless dies (red in
// the graveyard). It last existed colourless, so the copy MAY be
// pointed at the pro-red creature, and the copy then resolves there
// (its CR 608.2b re-check reads the same record, #1429).
func TestAbilityCopyOfADepartedColourlessSourceMayRetargetAProRedCreature(t *testing.T) {
	for _, kind := range []StackItemKind{StackItemTriggered, StackItemActivated} {
		t.Run(string(kind), func(t *testing.T) {
			s := newCopyRetargetScene(t, kind, false)
			s.kill(t, false)
			choice := s.copyWithNewTargets(t)

			if !slices.Contains(choice.PickTargetCards, s.proRed) {
				t.Error("the copy's re-target prompt does not offer the pro-red creature: the source " +
					"was COLOURLESS as it last existed (CR 707.10c / 608.2h) — it read the red graveyard card (#1449)")
			}
			s.refresh()
			if !slices.Contains(choice.PickTargetCards, s.proRed) {
				t.Error("the #809 refresh narrowed the prompt away from the pro-red creature: it read " +
					"the red graveyard card (#1449)")
			}
			if err := s.g.ResolvePickTargets(choice.ID, s.me.ID,
				[]TargetRef{{Kind: TargetCard, ID: s.proRed}}); err != nil {
				t.Fatalf("the answer refused the pro-red creature: %v (#1449)", err)
			}
			cp := abilityCopyOf(s.g, s.item.ID)
			if cp == nil || len(cp.Targets) != 1 || cp.Targets[0].ID != s.proRed {
				t.Fatalf("the copy should target the pro-red creature: %+v", cp)
			}
			if cp.SourceObject != s.item.SourceObject {
				t.Errorf("the copy carries its original's SourceObject (#1418): %+v, want %+v",
					cp.SourceObject, s.item.SourceObject)
			}
			resolveTop(t, s.g)
			if !*s.resolved {
				t.Error("the copy fizzled at the pro-red creature; its source last existed colourless")
			}
		})
	}
}

// A LIVE source is judged as it is now, exactly as before: the painted
// red permanent's copy may not take the pro-red creature, the bleached
// one's may.
func TestAbilityCopyOfALiveSourceRetargetsAsBefore(t *testing.T) {
	for _, kind := range []StackItemKind{StackItemTriggered, StackItemActivated} {
		for _, red := range []bool{true, false} {
			name := string(kind) + "/colourless"
			if red {
				name = string(kind) + "/red"
			}
			t.Run(name, func(t *testing.T) {
				s := newCopyRetargetScene(t, kind, red)
				choice := s.copyWithNewTargets(t)
				s.refresh()
				if got := slices.Contains(choice.PickTargetCards, s.proRed); got == red {
					t.Errorf("a live %s source: pro-red creature offered = %v, want %v", name, got, !red)
				}
				err := s.g.ResolvePickTargets(choice.ID, s.me.ID, []TargetRef{{Kind: TargetCard, ID: s.proRed}})
				if (err == nil) == red {
					t.Errorf("a live %s source: answering the pro-red creature = %v", name, err)
				}
			})
		}
	}
}

// A SPELL copy is judged against the spell, as before — even when the
// spell's card has a battlefield record this turn that the instance-ID
// rule would match. The state is #1429's hand-built one (a red card
// that was a bleached, colourless permanent, now on the stack with its
// epoch one past the record); no real path produces it, which is why
// the guard in copyTargetSourceLocked is a guard and not a branch.
func TestSpellCopyRetargetStillReadsTheSpell(t *testing.T) {
	g := newActiveGame(t)
	retargetSpecs(t)
	me, opp := g.Seats[0], g.Seats[1]
	bleached := map[uuid.UUID]bool{}
	withLKIColourPainters(t, nil, bleached)

	plain := pushColouredCreature(g, opp, "Plain", []string{"W"})
	proRed := pushColouredCreature(g, opp, "Pro-Red", []string{"W"}, "protection from red")
	lapsed := pushLKISource(g, me.ID, []string{"R"})
	bleached[lapsed] = true
	pushLKIPainter(g, me.ID, lkiBleacherOracle)
	destroy(t, g, lapsed)

	g.WithWriteLock(func() {
		c, err := me.Graveyard.Remove(lapsed)
		if err != nil {
			t.Fatalf("setup: the lapsed card is not in the graveyard: %v", err)
		}
		c.OracleID = retargetOneOracle
		c.TypeLine = "Instant"
		g.Stack.PushTop(c)
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		g.StackMeta[lapsed] = &StackItem{
			ID: lapsed, Kind: StackItemSpell, Controller: me.ID, Owner: me.ID, SourceCardID: lapsed,
			Targets: []TargetRef{{Kind: TargetCard, ID: plain}},
		}
		if rec, ok := g.departedDamageSourceLocked(lapsed, nil); !ok || len(rec.Characteristic.Colors) != 0 {
			t.Fatalf("setup: the instance-ID rule must match a colourless record for the "+
				"lapsed card, or this test proves nothing (ok=%v)", ok)
		}
		if err := g.CopySpellForEffect(lapsed, me.ID, true, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	choice := pendingOfKind(t, g, PendingChoicePickTarget)
	if !slices.Contains(choice.PickTargetCards, plain) {
		t.Fatalf("setup: the plain creature must be offered: %v", choice.PickTargetCards)
	}
	if slices.Contains(choice.PickTargetCards, proRed) {
		t.Error("a copy of a RED spell was offered the pro-red creature: it read the colourless " +
			"permanent the card was earlier in the turn")
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if slices.Contains(choice.PickTargetCards, proRed) {
		t.Error("the #809 refresh offered a red spell's copy the pro-red creature")
	}
	if err := g.ResolvePickTargets(choice.ID, me.ID, []TargetRef{{Kind: TargetCard, ID: proRed}}); err == nil {
		t.Error("a copy of a red spell was allowed to target the pro-red creature")
	}
}

// When every target the copy's prompt offered has gone by the #809
// refresh, the prompt is withdrawn and the copy is created keeping the
// original's targets (refreshTargetChoicesLocked). For an ABILITY copy
// that must be an ability copy: the withdrawal used to call the SPELL
// builder, which put a spell copy of the source permanent's card on the
// stack instead.
func TestWithdrawnAbilityCopyPromptCreatesAnAbilityCopy(t *testing.T) {
	s := newCopyRetargetScene(t, StackItemTriggered, false)
	choice := s.copyWithNewTargets(t)
	stackBefore := len(s.g.Stack.Cards)
	s.g.WithWriteLock(func() {
		for _, c := range slices.Clone(s.g.Battlefield.Cards) {
			if c.IsCreature() {
				if _, err := s.g.Battlefield.Remove(c.InstanceID); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}
		}
	})
	s.refresh()
	for _, pc := range s.g.PendingChoices {
		if pc.ID == choice.ID {
			t.Fatal("setup: the prompt should have been withdrawn with nothing left to offer")
		}
	}
	cp := abilityCopyOf(s.g, s.item.ID)
	if cp == nil {
		t.Fatal("withdrawing the prompt created no ability copy")
	}
	if cp.Kind != s.item.Kind || len(cp.Targets) != 1 || cp.Targets[0].ID != s.plain {
		t.Errorf("the copy should be a %s item keeping the original's target: %+v", s.item.Kind, cp)
	}
	if got := len(s.g.Stack.Cards); got != stackBefore {
		t.Errorf("the stack zone grew from %d to %d cards: a spell copy of the source permanent "+
			"was put on it", stackBefore, got)
	}
}

package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_onto_battlefield_test.go — #5931: a counter that puts the card
// onto the battlefield instead of into its owner's graveyard (Desertion,
// pinned end to end in cards/effects/desertion_test.go). Stubbed
// replacements, so these pin the engine contract: the entry window, the
// controller, the counter event, the cancel and the pause.

// entryOfSpell is a test replacement of the countered spell's entry onto
// the battlefield. `apply` is its Replace.
func entryOfSpell(owner, spell uuid.UUID, label string, apply func(ev *ReplacementEvent)) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == spell && ev.NewZone == ZoneBattlefield
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			apply(ev)
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      label,
	}
}

func counterOntoBattlefield(t *testing.T, g *Game, spell, controller uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CounterTargetToZoneForEffect(spell, ZoneRef{Kind: ZoneBattlefield, Owner: controller}); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
}

// The headline: the spell is countered (one EventCounterSpell, naming the
// countering player as its cause), leaves the stack with its record, and
// enters under the countering player's control. The entry window saw the
// new controller already stamped on the card.
func TestACounteredSpellEntersTheBattlefieldUnderTheCounterer(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushSpellFor(t, g, "Creature — Bear", me, me)
	var sawController uuid.UUID
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
				if ev.Kind == RepEventMove && ev.CardID == spell && ev.NewZone == ZoneBattlefield {
					if c, ok := g.LookupCardForEffect(spell); ok {
						sawController = c.Controller
					}
				}
				return false
			},
			Replace:    func(*ReplacementEvent, *Game, *Card) error { return nil },
			Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return me.ID },
			Label:      "observer",
		})
	})
	counteringFor(g, opp)
	counterOntoBattlefield(t, g, spell, opp.ID)

	c := findBattlefieldCard(g, spell)
	if c == nil || c.Controller != opp.ID || c.Owner != me.ID {
		t.Fatalf("permanent %+v: want it on the battlefield, controlled by the counterer, owned by the caster", c)
	}
	if g.Stack.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the card is still on the stack or also in its owner's graveyard")
	}
	if _, ok := g.StackMeta[spell]; ok {
		t.Error("the spell's stack record outlived it")
	}
	if sawController != opp.ID {
		t.Errorf("the entry window saw controller %v, want the counterer", sawController)
	}
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventCounterSpell && ev.CardID == spell {
			n++
			if ev.CauseController != opp.ID {
				t.Errorf("counter event cause controller %v, want the counterer", ev.CauseController)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d EventCounterSpell, want 1", n)
	}
	if !hasEventFor(g, EventETB, spell) {
		t.Error("no EventETB: the card entered without being announced")
	}
}

// An entry replacement applies to it: "enters tapped".
func TestACounteredSpellsEntryTakesEntryReplacements(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushSpellFor(t, g, "Artifact", me, me)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(entryOfSpell(me.ID, spell, "tapped", func(ev *ReplacementEvent) { ev.EntersTapped = true }))
	})
	counteringFor(g, opp)
	counterOntoBattlefield(t, g, spell, opp.ID)

	if c := findBattlefieldCard(g, spell); c == nil || !c.Tapped || c.Controller != opp.ID {
		t.Fatalf("permanent %+v: want it tapped under the counterer", c)
	}
}

// Cancelled, the entry does not happen, but the spell is still countered:
// it goes where CR 701.6a puts a countered spell, its owner's graveyard,
// with its controller put back.
func TestACancelledEntryStillCountersIntoTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushSpellFor(t, g, "Creature — Bear", me, me)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(entryOfSpell(me.ID, spell, "no entry", func(ev *ReplacementEvent) { ev.Canceled = true }))
	})
	counteringFor(g, opp)
	counterOntoBattlefield(t, g, spell, opp.ID)

	if g.Battlefield.Contains(spell) || g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Fatal("a cancelled entry did not leave the countered card in its owner's graveyard")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == spell && c.Controller == opp.ID {
			t.Error("the card kept the counterer as its controller in its owner's graveyard")
		}
	}
	if _, ok := g.StackMeta[spell]; ok {
		t.Error("the spell's stack record outlived it")
	}
	if !hasEventFor(g, EventCounterSpell, spell) {
		t.Error("the spell was not countered")
	}
}

// The entry can pause on a CR 616 ordering prompt. Nothing has moved
// while it is open, the spell is still on the stack, and the answer
// finishes the entry and the counter.
func TestACounteredSpellsEntryPausesAndResumes(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := pushSpellFor(t, g, "Creature — Bear", me, me)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(entryOfSpell(me.ID, spell, "tapped", func(ev *ReplacementEvent) { ev.EntersTapped = true }))
		g.RegisterReplacementForTest(entryOfSpell(me.ID, spell, "a counter", func(ev *ReplacementEvent) { ev.AddCounterAtETB("+1/+1", 1) }))
	})
	counteringFor(g, opp)
	counterOntoBattlefield(t, g, spell, opp.ID)

	p := onlyPrompt(t, g)
	if p.Kind != PendingChoiceReplacementOrder {
		t.Fatalf("prompt kind = %q, want %q", p.Kind, PendingChoiceReplacementOrder)
	}
	if !g.Stack.Contains(spell) || hasEventFor(g, EventCounterSpell, spell) {
		t.Fatal("the card moved, or was reported countered, while the prompt was open")
	}
	if err := g.ResolveReplacementOrder(p.ID, p.Chooser, p.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	c := findBattlefieldCard(g, spell)
	if c == nil || c.Controller != opp.ID || !c.Tapped || c.Counters["+1/+1"] != 1 {
		t.Fatalf("permanent %+v: want it tapped with a counter under the counterer", c)
	}
	if !hasEventFor(g, EventCounterSpell, spell) {
		t.Error("the resumed entry did not report the counter")
	}
	if _, ok := g.StackMeta[spell]; ok {
		t.Error("the spell's stack record outlived it")
	}
}

// The battlefield is a destination for a COUNTER only: a plain exit from
// the stack onto it is refused, with the spell left where it was.
func TestOnlyACounterMayPutASpellOntoTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := pushSpellFor(t, g, "Creature — Bear", me, me)
	var err error
	g.WithWriteLock(func() {
		err = g.exitSpellFromStackLocked(spell, &ZoneRef{Kind: ZoneBattlefield}, ZoneHand, false)
	})
	if !errors.Is(err, ErrInvalidStackDestination) {
		t.Errorf("err = %v, want ErrInvalidStackDestination", err)
	}
	if !g.Stack.Contains(spell) {
		t.Error("a refused exit moved the spell")
	}
}

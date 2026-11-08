package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// counter_unless_paid_dest_test.go — #5236: a counter-unless-pays whose
// countered spell goes somewhere other than its owner's graveyard
// (Syncopate's and No More Lies's exile). The destination rides the
// prompt as the decline's key, so it needs no field of its own.

func queueCounterTaxTo(g *Game, spell uuid.UUID, dst ZoneRef, then PayUnlessElse) error {
	var err error
	g.WithWriteLock(func() {
		err = g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: spell,
			Source:    uuid.New(),
			Cost:      "{2}",
			Question:  "pay {2} or the spell is exiled",
			Dest:      dst,
			Then:      then,
		})
	})
	return err
}

// Exile is keyed, and a decline exiles the spell instead of putting it
// into its owner's graveyard.
func TestACounterTaxToExileExilesOnADecline(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	if err := queueCounterTaxTo(g, spell, ZoneRef{Kind: ZoneExile}, PayUnlessElse{}); err != nil {
		t.Fatalf("queue: %v", err)
	}
	prompt := onlyPendingChoice(t, g)
	if prompt.PickThen != counterGuardedExileDecline.Key() || prompt.GuardsStackItem != spell {
		t.Fatalf("prompt key %q guards %s, want %q guarding the spell",
			prompt.PickThen, prompt.GuardsStackItem, counterGuardedExileDecline.Key())
	}
	if err := g.ResolvePayUnless(prompt.ID, caster.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if g.Stack.Contains(spell) || !g.Exile.Contains(spell) || caster.Graveyard.Contains(spell) {
		t.Errorf("on stack=%v exiled=%v in graveyard=%v, want it exiled",
			g.Stack.Contains(spell), g.Exile.Contains(spell), caster.Graveyard.Contains(spell))
	}
}

// The zero destination is the historic key, so every existing
// counter-unless-pays queues exactly the prompt it did.
func TestACounterTaxWithNoDestinationKeepsTheGraveyardKey(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	queueWardTax(t, g, spell, uuid.Nil)
	prompt := onlyPendingChoice(t, g)
	if prompt.PickThen != "pay-unless/counter-guarded" {
		t.Fatalf("prompt key %q, want the historic pay-unless/counter-guarded", prompt.PickThen)
	}
	if err := g.ResolvePayUnless(prompt.ID, caster.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if g.Exile.Contains(spell) || !caster.Graveyard.Contains(spell) {
		t.Errorf("exiled=%v in graveyard=%v, want it in its owner's graveyard",
			g.Exile.Contains(spell), caster.Graveyard.Contains(spell))
	}
}

// A destination with no keyed decline, or one beside a card's own
// branches, is refused and asks nothing.
func TestACounterTaxRefusesADestinationItCannotKey(t *testing.T) {
	for _, tc := range []struct {
		name string
		dst  ZoneRef
		then PayUnlessElse
	}{
		{"hand", ZoneRef{Kind: ZoneHand}, PayUnlessElse{}},
		{"a named player's exile", ZoneRef{Kind: ZoneExile, Owner: uuid.New()}, PayUnlessElse{}},
		{"beside branches", ZoneRef{Kind: ZoneExile}, counterGuardedDecline},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			spell := guardedSpellOnStack(t, g, g.Seats[1])
			if err := queueCounterTaxTo(g, spell, tc.dst, tc.then); !errors.Is(err, ErrInvalidParam) {
				t.Fatalf("queue = %v, want ErrInvalidParam", err)
			}
			if len(g.PendingChoices) != 0 {
				t.Errorf("%d prompts queued for a refused destination", len(g.PendingChoices))
			}
		})
	}
}

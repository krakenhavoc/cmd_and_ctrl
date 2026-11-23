package testsp

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An opponent's discard: it goes to the graveyard and comes straight back
// to its owner's hand as the trigger resolves.
func TestQuagnothReturnsToHandWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	q := d.HandCard(d.Victim, "Quagnoth", "Creature — Beast", "{5}{G}", effectstest.QuagnothOracle, 4, 5)
	filler := d.Filler(d.Victim)
	w := d.Watch()

	d.TheyDiscard(t, q, filler)

	if !d.Victim.Hand.Contains(q) {
		t.Errorf("Quagnoth is in %q, want the hand", d.ZoneOf(q))
	}
	if !d.Victim.Graveyard.Contains(filler) {
		t.Error("the other card did not stay in the graveyard")
	}
	var discarded bool
	for _, ev := range w.Events {
		if ev.CardID == q {
			discarded = ev.NewZone == game.ZoneGraveyard
		}
	}
	if !discarded {
		t.Error("Quagnoth's discard did not land in the graveyard first")
	}
}

func TestQuagnothStaysInTheGraveyardWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	q := d.HandCard(d.Caster, "Quagnoth", "Creature — Beast", "{5}{G}", effectstest.QuagnothOracle, 4, 5)
	filler := d.Filler(d.Caster)

	d.IDiscard(t, q, filler)

	if !d.Caster.Graveyard.Contains(q) {
		t.Errorf("Quagnoth is in %q, want the graveyard", d.ZoneOf(q))
	}
}

func TestQuagnothPrintedKeywords(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.QuagnothOracle)
	if !ok {
		t.Fatal("Quagnoth is not registered")
	}
	want := map[string]bool{"split second": true, "shroud": true}
	for _, k := range spec.PrintedKeywords {
		delete(want, k)
	}
	if len(want) != 0 {
		t.Errorf("missing printed keywords %v", want)
	}
}

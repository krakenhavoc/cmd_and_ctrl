package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// quagnoth_test.go — ADR 0153 PR 2 (#3168).

const quagnothOracle = "9250eb9e-690f-4563-8132-d737f6d22883"

// An opponent's discard: it goes to the graveyard and comes straight back
// to its owner's hand as the trigger resolves.
func TestQuagnothReturnsToHandWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	q := d.handCard(d.victim, "Quagnoth", "Creature — Beast", "{5}{G}", quagnothOracle, 4, 5)
	filler := d.filler(d.victim)
	w := d.watch()

	d.theyDiscard(t, q, filler)

	if !d.victim.Hand.Contains(q) {
		t.Errorf("Quagnoth is in %q, want the hand", d.zoneOf(q))
	}
	if !d.victim.Graveyard.Contains(filler) {
		t.Error("the other card did not stay in the graveyard")
	}
	var discarded bool
	for _, ev := range w.events {
		if ev.CardID == q {
			discarded = ev.NewZone == game.ZoneGraveyard
		}
	}
	if !discarded {
		t.Error("Quagnoth's discard did not land in the graveyard first")
	}
}

func TestQuagnothStaysInTheGraveyardWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	q := d.handCard(d.caster, "Quagnoth", "Creature — Beast", "{5}{G}", quagnothOracle, 4, 5)
	filler := d.filler(d.caster)

	d.iDiscard(t, q, filler)

	if !d.caster.Graveyard.Contains(q) {
		t.Errorf("Quagnoth is in %q, want the graveyard", d.zoneOf(q))
	}
}

func TestQuagnothPrintedKeywords(t *testing.T) {
	spec, ok := Lookup(quagnothOracle)
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

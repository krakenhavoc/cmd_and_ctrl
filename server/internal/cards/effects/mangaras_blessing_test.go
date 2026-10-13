package effects

import "testing"

// mangaras_blessing_test.go — ADR 0153 PR 2 (#3168).

const mangarasBlessingOracle = "5d9431de-7cf0-43d4-9b86-e0ef4a7756f9"

// An opponent's discard: 2 life now, and the card returns to its owner's
// hand at the beginning of the next end step.
func TestMangarasBlessingGainsTwoAndReturnsAtTheNextEndStep(t *testing.T) {
	d := newDBOTable(t)
	card := d.handCard(d.victim, "Mangara's Blessing", "Instant", "{2}{W}", mangarasBlessingOracle, 0, 0)
	filler := d.filler(d.victim)
	life := d.victim.Life

	d.theyDiscard(t, card, filler)

	if d.victim.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, d.victim.Life)
	}
	if !d.victim.Graveyard.Contains(card) {
		t.Fatalf("before the end step the card is in %q, want the graveyard", d.zoneOf(card))
	}
	d.toEndStep(t)
	if !d.victim.Hand.Contains(card) {
		t.Errorf("at the end step the card is in %q, want the hand", d.zoneOf(card))
	}
	if !d.victim.Graveyard.Contains(filler) {
		t.Error("the other discarded card left the graveyard")
	}
}

// A card that has left the graveyard by the end step is not returned.
func TestMangarasBlessingDoesNotReturnACardThatLeftTheGraveyard(t *testing.T) {
	d := newDBOTable(t)
	card := d.handCard(d.victim, "Mangara's Blessing", "Instant", "{2}{W}", mangarasBlessingOracle, 0, 0)
	filler := d.filler(d.victim)

	d.theyDiscard(t, card, filler)
	d.g.WithWriteLock(func() {
		if err := d.g.ExileCardForEffect(card); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	d.toEndStep(t)

	if d.victim.Hand.Contains(card) {
		t.Error("the exiled card came back to the hand")
	}
}

func TestMangarasBlessingDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	card := d.handCard(d.caster, "Mangara's Blessing", "Instant", "{2}{W}", mangarasBlessingOracle, 0, 0)
	filler := d.filler(d.caster)
	life := d.caster.Life

	d.iDiscard(t, card, filler)
	d.toEndStep(t)

	if d.caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.caster.Life)
	}
	if !d.caster.Graveyard.Contains(card) {
		t.Errorf("the card is in %q, want the graveyard", d.zoneOf(card))
	}
}

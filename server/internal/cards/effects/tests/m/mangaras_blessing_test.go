package testsm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// An opponent's discard: 2 life now, and the card returns to its owner's
// hand at the beginning of the next end step.
func TestMangarasBlessingGainsTwoAndReturnsAtTheNextEndStep(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	card := d.HandCard(d.Victim, "Mangara's Blessing", "Instant", "{2}{W}", effectstest.MangarasBlessingOracle, 0, 0)
	filler := d.Filler(d.Victim)
	life := d.Victim.Life

	d.TheyDiscard(t, card, filler)

	if d.Victim.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, d.Victim.Life)
	}
	if !d.Victim.Graveyard.Contains(card) {
		t.Fatalf("before the end step the card is in %q, want the graveyard", d.ZoneOf(card))
	}
	d.ToEndStep(t)
	if !d.Victim.Hand.Contains(card) {
		t.Errorf("at the end step the card is in %q, want the hand", d.ZoneOf(card))
	}
	if !d.Victim.Graveyard.Contains(filler) {
		t.Error("the other discarded card left the graveyard")
	}
}

// A card that has left the graveyard by the end step is not returned.
func TestMangarasBlessingDoesNotReturnACardThatLeftTheGraveyard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	card := d.HandCard(d.Victim, "Mangara's Blessing", "Instant", "{2}{W}", effectstest.MangarasBlessingOracle, 0, 0)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, card, filler)
	d.G.WithWriteLock(func() {
		if err := d.G.ExileCardForEffect(card); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	d.ToEndStep(t)

	if d.Victim.Hand.Contains(card) {
		t.Error("the exiled card came back to the hand")
	}
}

func TestMangarasBlessingDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	card := d.HandCard(d.Caster, "Mangara's Blessing", "Instant", "{2}{W}", effectstest.MangarasBlessingOracle, 0, 0)
	filler := d.Filler(d.Caster)
	life := d.Caster.Life

	d.IDiscard(t, card, filler)
	d.ToEndStep(t)

	if d.Caster.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.Caster.Life)
	}
	if !d.Caster.Graveyard.Contains(card) {
		t.Errorf("the card is in %q, want the graveyard", d.ZoneOf(card))
	}
}

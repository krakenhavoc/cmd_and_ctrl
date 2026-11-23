package testsl

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loxodon_smiter_test.go — ADR 0153 PR 2 (#3168).

const loxodonSmiterOracle = "18ac2282-83af-4acc-b20b-6a95771da688"

func TestLoxodonSmiterEntersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	smiter := d.HandCard(d.Victim, "Loxodon Smiter", "Creature — Elephant Soldier", "{1}{G}{W}", loxodonSmiterOracle, 4, 4)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, smiter, filler)

	c, ok := d.OnBattlefield(smiter)
	if !ok {
		t.Fatalf("the Smiter is in %q, want the battlefield", d.ZoneOf(smiter))
	}
	if c.Controller != d.Victim.ID {
		t.Errorf("the Smiter is controlled by %v, want its owner", c.Controller)
	}
	if d.Victim.Graveyard.Contains(smiter) {
		t.Error("the Smiter is also in the graveyard")
	}
}

func TestLoxodonSmiterStaysInTheGraveyardForYourOwnDiscard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	smiter := d.HandCard(d.Caster, "Loxodon Smiter", "Creature — Elephant Soldier", "{1}{G}{W}", loxodonSmiterOracle, 4, 4)
	filler := d.Filler(d.Caster)

	d.IDiscard(t, smiter, filler)

	if !d.Caster.Graveyard.Contains(smiter) {
		t.Fatalf("the Smiter is in %q, want the graveyard", d.ZoneOf(smiter))
	}
}

// "This spell can't be countered" is the first printed line.
func TestLoxodonSmiterCantBeCountered(t *testing.T) {
	spec, ok := effects.Lookup(loxodonSmiterOracle)
	if !ok {
		t.Fatal("Loxodon Smiter is not registered")
	}
	if !spec.CantBeCountered {
		t.Error("Loxodon Smiter does not declare that it can't be countered")
	}
	if spec.DiscardByOpponent != game.DiscardByOpponentBattlefield {
		t.Errorf("DiscardByOpponent = %q, want battlefield", spec.DiscardByOpponent)
	}
}

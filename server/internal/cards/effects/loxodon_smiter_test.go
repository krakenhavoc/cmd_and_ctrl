package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loxodon_smiter_test.go — ADR 0153 PR 2 (#3168).

const loxodonSmiterOracle = "18ac2282-83af-4acc-b20b-6a95771da688"

func TestLoxodonSmiterEntersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	smiter := d.handCard(d.victim, "Loxodon Smiter", "Creature — Elephant Soldier", "{1}{G}{W}", loxodonSmiterOracle, 4, 4)
	filler := d.filler(d.victim)

	d.theyDiscard(t, smiter, filler)

	c, ok := d.onBattlefield(smiter)
	if !ok {
		t.Fatalf("the Smiter is in %q, want the battlefield", d.zoneOf(smiter))
	}
	if c.Controller != d.victim.ID {
		t.Errorf("the Smiter is controlled by %v, want its owner", c.Controller)
	}
	if d.victim.Graveyard.Contains(smiter) {
		t.Error("the Smiter is also in the graveyard")
	}
}

func TestLoxodonSmiterStaysInTheGraveyardForYourOwnDiscard(t *testing.T) {
	d := newDBOTable(t)
	smiter := d.handCard(d.caster, "Loxodon Smiter", "Creature — Elephant Soldier", "{1}{G}{W}", loxodonSmiterOracle, 4, 4)
	filler := d.filler(d.caster)

	d.iDiscard(t, smiter, filler)

	if !d.caster.Graveyard.Contains(smiter) {
		t.Fatalf("the Smiter is in %q, want the graveyard", d.zoneOf(smiter))
	}
}

// "This spell can't be countered" is the first printed line.
func TestLoxodonSmiterCantBeCountered(t *testing.T) {
	spec, ok := Lookup(loxodonSmiterOracle)
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

package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guerrilla_tactics_test.go — ADR 0153 PR 2 (#3168).

const guerrillaTacticsOracle = "65ad90f5-4dae-494b-83fb-8052e77e0b12"

func TestGuerrillaTacticsCastDealsTwo(t *testing.T) {
	d := newDBOTable(t)
	life := d.victim.Life

	castCatalogSpell(t, d.g, "Guerrilla Tactics", "Instant", guerrillaTacticsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)

	if d.victim.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, d.victim.Life)
	}
}

// An opponent's discard: 4 damage to any target, chosen by the card's
// owner as the ability goes on the stack.
func TestGuerrillaTacticsDealsFourWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	g := d.handCard(d.victim, "Guerrilla Tactics", "Instant", "{1}{R}", guerrillaTacticsOracle, 0, 0)
	filler := d.filler(d.victim)
	life := d.caster.Life

	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, g, filler)
	passPriorityAroundTable(t, d.g)
	pickTargetRefs(t, d.g, d.victim.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: d.caster.ID})
	passPriorityAroundTable(t, d.g)

	if d.caster.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, d.caster.Life)
	}
	if !d.victim.Graveyard.Contains(g) {
		t.Errorf("the card is in %q, want the graveyard", d.zoneOf(g))
	}
}

func TestGuerrillaTacticsDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	g := d.handCard(d.caster, "Guerrilla Tactics", "Instant", "{1}{R}", guerrillaTacticsOracle, 0, 0)
	filler := d.filler(d.caster)
	life := d.victim.Life

	d.iDiscard(t, g, filler)

	if latestPickTarget(d.g, d.caster.ID) != nil {
		t.Error("your own discard put the trigger on the stack")
	}
	if d.victim.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.victim.Life)
	}
}

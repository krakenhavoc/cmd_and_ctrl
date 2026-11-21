package testsg

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestGuerrillaTacticsCastDealsTwo(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	life := d.Victim.Life

	effectstest.CastCatalogSpell(t, d.G, "Guerrilla Tactics", "Instant", effectstest.GuerrillaTacticsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)

	if d.Victim.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, d.Victim.Life)
	}
}

// An opponent's discard: 4 damage to any target, chosen by the card's
// owner as the ability goes on the stack.
func TestGuerrillaTacticsDealsFourWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	g := d.HandCard(d.Victim, "Guerrilla Tactics", "Instant", "{1}{R}", effectstest.GuerrillaTacticsOracle, 0, 0)
	filler := d.Filler(d.Victim)
	life := d.Caster.Life

	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, g, filler)
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.PickTargetRefs(t, d.G, d.Victim.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: d.Caster.ID})
	effectstest.PassPriorityAroundTable(t, d.G)

	if d.Caster.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, d.Caster.Life)
	}
	if !d.Victim.Graveyard.Contains(g) {
		t.Errorf("the card is in %q, want the graveyard", d.ZoneOf(g))
	}
}

func TestGuerrillaTacticsDoesNothingWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	g := d.HandCard(d.Caster, "Guerrilla Tactics", "Instant", "{1}{R}", effectstest.GuerrillaTacticsOracle, 0, 0)
	filler := d.Filler(d.Caster)
	life := d.Victim.Life

	d.IDiscard(t, g, filler)

	if effectstest.LatestPickTarget(d.G, d.Caster.ID) != nil {
		t.Error("your own discard put the trigger on the stack")
	}
	if d.Victim.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, d.Victim.Life)
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB162ShatterOracle = "c1372f18-9b33-40e8-bfe2-dc4ca3b3ef55"

func oldB162ShatterHand(opp *game.Player) (rock, land, mono, gold uuid.UUID) {
	rock = effectstest.HandCardFull(opp, "Their Rock", "Artifact", "{2}", "", nil)
	land = effectstest.HandCardFull(opp, "Their Wastes", "Basic Land", "", "", nil)
	mono = effectstest.HandCardFull(opp, "Their Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	gold = effectstest.HandCardFull(opp, "Their Gold Card", "Creature — Elf", "{G}{W}", "", []string{"G", "W"})
	return
}

func TestOldB162ShatterAssumptionsColorlessModeDiscardsColorlessNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock, land, mono, gold := oldB162ShatterHand(opp)
	effectstest.P1g3CastModal(t, g, "Shatter Assumptions", "Sorcery", oldB162ShatterOracle, []int{0}, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the colorless nonland card was not discarded")
	}
	for _, id := range []uuid.UUID{land, mono, gold} {
		if !opp.Hand.Contains(id) {
			t.Errorf("a card that is not colorless nonland was discarded: %q", effectstest.ZoneOf(g, id))
		}
	}
}

func TestOldB162ShatterAssumptionsMulticoloredModeDiscardsOnlyMulticoloredCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock, land, mono, gold := oldB162ShatterHand(opp)
	effectstest.P1g3CastModal(t, g, "Shatter Assumptions", "Sorcery", oldB162ShatterOracle, []int{1}, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(gold) || !opp.Graveyard.Contains(gold) {
		t.Error("the multicolored card was not discarded")
	}
	for _, id := range []uuid.UUID{rock, land, mono} {
		if !opp.Hand.Contains(id) {
			t.Errorf("a card that is not multicolored was discarded: %q", effectstest.ZoneOf(g, id))
		}
	}
}

func TestOldB162ShatterAssumptionsRefusesAnUntargetablePlayerSelf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	card := effectstest.HandCardFull(me, "Shatter Assumptions", "Sorcery", "{1}{B}{B}", oldB162ShatterOracle, []string{"B"})
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("targeted yourself with \"target opponent\"")
	}
}

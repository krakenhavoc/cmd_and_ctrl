package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB162RageOracle = "c019ed4d-33a3-48be-a38e-93e5d6161b1f"

func oldB162Elementals(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Name == "Elemental" {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB162ForceOfRagePitchesOffTheirTurnAndSacrificesAtYourNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	red := effectstest.HandCardFull(opp, "Their Red Card", "Instant", "{R}", "", []string{"R"})
	green := effectstest.HandCardFull(opp, "Their Green Card", "Instant", "{G}", "", []string{"G"})
	rage := effectstest.HandCardFull(opp, "Force of Rage", "Instant", "{1}{R}{R}", oldB162RageOracle, []string{"R"})
	if err := g.CastSpell(opp.ID, rage, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{green}}); err == nil {
		t.Fatal("a green card paid for the red pitch")
	}
	if err := g.CastSpell(opp.ID, rage, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{red}}); err != nil {
		t.Fatalf("Force of Rage off the pitch on another player's turn: %v", err)
	}
	if opp.Hand.Contains(red) || !g.Exile.Contains(red) {
		t.Error("the pitched red card must be exiled")
	}
	effectstest.PassPriorityAroundTable(t, g)
	els := oldB162Elementals(g, opp.ID)
	if len(els) != 2 {
		t.Fatalf("%d Elementals, want 2", len(els))
	}
	for _, c := range els {
		if c.Power != 3 || c.Toughness != 1 || !game.HasKeyword(&c, "trample") || !game.HasKeyword(&c, "haste") {
			t.Errorf("Elemental is %d/%d (trample %v, haste %v), want a 3/1 with trample and haste", c.Power, c.Toughness, game.HasKeyword(&c, "trample"), game.HasKeyword(&c, "haste"))
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB162Elementals(g, opp.ID)); n != 0 {
		t.Errorf("%d Elementals survived their controller's next upkeep", n)
	}
}

func TestOldB162ForceOfRageCannotPitchOnYourOwnTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	red := effectstest.HandCardFull(me, "My Red Card", "Instant", "{R}", "", []string{"R"})
	rage := effectstest.HandCardFull(me, "Force of Rage", "Instant", "{1}{R}{R}", oldB162RageOracle, []string{"R"})
	if err := g.CastSpell(me.ID, rage, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{red}}); err == nil {
		t.Fatal("pitched Force of Rage on your own turn")
	}
}

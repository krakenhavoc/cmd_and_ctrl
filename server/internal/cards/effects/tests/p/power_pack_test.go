package testsp

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Power Pack's combat damage exiles an instant or sorcery card from its
// controller's graveyard drawn at random — never another kind of card.
func TestPowerPackExilesAnInstantOrSorceryChosenAtRandom(t *testing.T) {
	seen := map[string]bool{}
	for key := byte(0); key < 48 && len(seen) < 2; key++ {
		g := effectstest.NewCatalogGame(t)
		g.SetRNGKeyForTest(effectstest.RandomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		pp := effectstest.B12Push(g, me.ID, "Power Pack", "Legendary Creature — Human Hero", effectstest.PowerPackOracle, 4, 4)
		shock := effectstest.PpGraveyardCard(g, me, "Shock", "Instant", "")
		div := effectstest.PpGraveyardCard(g, me, "Divination", "Sorcery", effectstest.SpellControlDivinationOracle)
		bearCard := effectstest.PpGraveyardCard(g, me, "Bear Card", "Creature — Bear", "")
		effectstest.DealCombatDamageToPlayer(g, pp, opp.ID, 4)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			t.Fatalf("a target chosen at random asks nobody: %+v", p)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.InExile(g, bearCard) {
			t.Fatalf("key %d: a creature card was exiled", key)
		}
		switch {
		case effectstest.InExile(g, shock) && !effectstest.InExile(g, div):
			seen["instant"] = true
		case effectstest.InExile(g, div) && !effectstest.InExile(g, shock):
			seen["sorcery"] = true
		default:
			t.Fatalf("key %d: want exactly one of the instant and the sorcery exiled", key)
		}
	}
	if len(seen) != 2 {
		t.Errorf("48 draws reached %v; both cards qualify", seen)
	}
}

// At your next upkeep the exiled card may be cast free, and it is
// exiled instead of going to the graveyard.
func TestPowerPackCastsTheExiledCardFreeAtYourNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pp := effectstest.B12Push(g, me.ID, "Power Pack", "Legendary Creature — Human Hero", effectstest.PowerPackOracle, 4, 4)
	div := effectstest.PpGraveyardCard(g, me, "Divination", "Sorcery", effectstest.SpellControlDivinationOracle)
	effectstest.DealCombatDamageToPlayer(g, pp, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, div) {
		t.Fatal("Divination was not exiled")
	}
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Fatal("the card was castable before your next upkeep")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the exiled card for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after Divination, want %d", got, hand+2)
	}
	if me.Graveyard.Contains(div) || !effectstest.InExile(g, div) {
		t.Error("the spell went to the graveyard instead of exile")
	}
}

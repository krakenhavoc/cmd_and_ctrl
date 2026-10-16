package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// power_pack_test.go — Power Pack (#1815): a graveyard target chosen at
// random, then a free cast at your next upkeep.

const powerPackOracle = "3d5c98ff-fcf1-42bd-9533-2648791a4f45"

func ppGraveyardCard(g *game.Game, owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Owner: owner.ID, Controller: owner.ID})
	})
	return id
}

// Power Pack's combat damage exiles an instant or sorcery card from its
// controller's graveyard drawn at random — never another kind of card.
func TestPowerPackExilesAnInstantOrSorceryChosenAtRandom(t *testing.T) {
	seen := map[string]bool{}
	for key := byte(0); key < 48 && len(seen) < 2; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		pp := b12Push(g, me.ID, "Power Pack", "Legendary Creature — Human Hero", powerPackOracle, 4, 4)
		shock := ppGraveyardCard(g, me, "Shock", "Instant", "")
		div := ppGraveyardCard(g, me, "Divination", "Sorcery", spellControlDivinationOracle)
		bearCard := ppGraveyardCard(g, me, "Bear Card", "Creature — Bear", "")
		dealCombatDamageToPlayer(g, pp, opp.ID, 4)
		if p := latestPickTarget(g, me.ID); p != nil {
			t.Fatalf("a target chosen at random asks nobody: %+v", p)
		}
		passPriorityAroundTable(t, g)
		if inExile(g, bearCard) {
			t.Fatalf("key %d: a creature card was exiled", key)
		}
		switch {
		case inExile(g, shock) && !inExile(g, div):
			seen["instant"] = true
		case inExile(g, div) && !inExile(g, shock):
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pp := b12Push(g, me.ID, "Power Pack", "Legendary Creature — Human Hero", powerPackOracle, 4, 4)
	div := ppGraveyardCard(g, me, "Divination", "Sorcery", spellControlDivinationOracle)
	dealCombatDamageToPlayer(g, pp, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if !inExile(g, div) {
		t.Fatal("Divination was not exiled")
	}
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err == nil {
		t.Fatal("the card was castable before your next upkeep")
	}
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the exiled card for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after Divination, want %d", got, hand+2)
	}
	if me.Graveyard.Contains(div) || !inExile(g, div) {
		t.Error("the spell went to the graveyard instead of exile")
	}
}

package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const seasonOfGrowthOracle = "3f4e300a-ec5c-42f3-a97b-d58e62abe22b"

// TestSeasonOfGrowthScriesWhenYourCreatureEnters.
func TestSeasonOfGrowthScriesWhenYourCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom Me")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Season of Growth", TypeLine: "Enchantment",
		OracleID: seasonOfGrowthOracle, Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "season-bear-test", nil)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil {
		t.Fatalf("a creature you control entering should scry 1: %+v", g.PendingChoices)
	}
}

// TestSeasonOfGrowthDrawsOnceWhenASpellTargetsYourCreature, even when
// the spell targets your creature twice.
func TestSeasonOfGrowthDrawsOnceWhenASpellTargetsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Season of Growth", TypeLine: "Enchantment",
		OracleID: seasonOfGrowthOracle, Owner: me.ID, Controller: me.ID,
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", "gg-test",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+1 {
		t.Errorf("hand after casting a spell targeting your creature: %d, want %d", me.Hand.Size(), hand+1)
	}
}

// TestSeasonOfGrowthDoesNotDrawForAnOpponentsCreature.
func TestSeasonOfGrowthDoesNotDrawForAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Season of Growth", TypeLine: "Enchantment",
		OracleID: seasonOfGrowthOracle, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Doom Blade", "Instant", "doomblade-test",
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	if me.Hand.Size() != hand {
		t.Errorf("a spell targeting an opponent's creature should not draw: %d, want %d", me.Hand.Size(), hand)
	}
}

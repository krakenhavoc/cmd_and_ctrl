package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const thassaGodOfTheSeaOracle = "69e5df2f-be1f-4608-a90f-3e2f51e2fea4"

// TestThassaIsNotACreatureUnderFiveDevotion checks the layer-4 clause
// and that indestructible is present regardless of devotion.
func TestThassaIsNotACreatureUnderFiveDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thassa := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thassa, God of the Sea",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: thassaGodOfTheSeaOracle,
		ManaCost: "{2}{U}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})

	types := effectstest.EffectiveTypes(t, g, thassa)
	if effectstest.ContainsStr(types, "Creature") {
		t.Errorf("with only Thassa's own devotion (1), she should not be a creature: %v", types)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, thassa, "indestructible") {
		t.Error("Thassa should always have indestructible")
	}

	// Four more blue pips of devotion (five total with Thassa's own)
	// turn her into a creature.
	for i := 0; i < 4; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Blue Permanent", TypeLine: "Artifact",
			ManaCost: "{U}", Owner: me.ID, Controller: me.ID,
		})
	}
	types = effectstest.EffectiveTypes(t, g, thassa)
	if !effectstest.ContainsStr(types, "Creature") {
		t.Errorf("at devotion 5, Thassa should be a creature: %v", types)
	}
}

// TestThassaUpkeepScries1.
func TestThassaUpkeepScries1(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom Me")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thassa, God of the Sea",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: thassaGodOfTheSeaOracle,
		ManaCost: "{2}{U}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})

	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil {
		t.Fatalf("no scry-1 prompt from Thassa's upkeep trigger: %+v", g.PendingChoices)
	}
}

// TestThassaMakesYourCreatureUnblockable — the {1}{U} ability is
// scoped to creatures the controller controls.
func TestThassaMakesYourCreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thassa := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thassa, God of the Sea",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: thassaGodOfTheSeaOracle,
		ManaCost: "{2}{U}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, thassa, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertRestrictions(t, g, mine, game.CantBeBlocked)
}

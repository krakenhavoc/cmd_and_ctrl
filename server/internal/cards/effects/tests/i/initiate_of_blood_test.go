package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func initiateOfBloodRow() cards.Card {
	return effectstest.FlipRow(effects.InitiateOfBloodOracleID, "{3}{R}", []string{"R"}, nil,
		cards.CardFace{Name: "Initiate of Blood", TypeLine: "Creature — Ogre Shaman",
			OracleText: "{T}: This creature deals 1 damage to target creature that was dealt damage this turn. When that creature dies this turn, flip this creature.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Goka the Unjust", TypeLine: "Legendary Creature — Ogre Shaman",
			OracleText: "{T}: Goka deals 4 damage to target creature that was dealt damage this turn.",
			Power:      "4", Toughness: "4"})
}

func initiateTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := effectstest.FlipMain(t)
	return g, me, opp, effectstest.FlipEntered(t, g, initiateOfBloodRow(), me)
}

// The Initiate can only aim at a creature that was dealt damage this
// turn.
func TestInitiateOfBloodNeedsADamagedCreature(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 3, 3))
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("targeted a creature that was not dealt damage this turn")
	}
	effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("a damaged creature is a legal target: %v", err)
	}
}

// The damage kills the creature it was aimed at: "when that creature
// dies this turn" flips the Initiate, as the card intends.
func TestInitiateOfBloodFlipsWhenItsDamageKillsTheCreature(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 2, 2))
	effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if effects.OnBattlefield(g, bear) {
		t.Fatal("the bear survived 2 damage")
	}
	effectstest.WantFlipped(t, g, id, "Goka the Unjust", []string{"R"}, 4)
}

// A creature that survives only flips the Initiate if it dies later
// the same turn, and not on a later turn.
func TestInitiateOfBloodFlipsOnlyIfThatCreatureDiesThisTurn(t *testing.T) {
	t.Run("dies later this turn", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 3, 3))
		effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
		effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
		effectstest.WantUnflipped(t, g, id, "Initiate of Blood")
		effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.WantFlipped(t, g, id, "Goka the Unjust", []string{"R"}, 4)
	})
	t.Run("dies next turn", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 3, 3))
		effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
		effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
		effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
		effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 5)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.WantUnflipped(t, g, id, "Initiate of Blood")
	})
	t.Run("a different creature dies", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 3, 3))
		other := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Other", "oracle-other", 1, 1))
		effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
		effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
		effectstest.DamageAndSettle(t, g, uuid.Nil, other, 1)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.WantUnflipped(t, g, id, "Initiate of Blood")
	})
	t.Run("the Initiate is gone", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 3, 3))
		effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
		effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 5)
		effectstest.PassPriorityAroundTable(t, g)
		if c, ok := g.LookupCardForEffect(id); !ok || c.Flipped || c.ActiveFace != 0 {
			t.Errorf("the dead Initiate is flipped %v face %d", c.Flipped, c.ActiveFace)
		}
	})
}

// A returned Initiate is a new object and the old trigger flips
// nothing (CR 400.7).
func TestInitiateOfBloodDoesNotFlipANewObject(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 3, 3))
	effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	g.WithWriteLock(func() {
		if err := (effects.BounceToHand{Target: id}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
		if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantUnflipped(t, g, id, "Initiate of Blood")
}

// Goka deals 4 to a creature that was dealt damage this turn.
func TestGokaDealsFourToADamagedCreature(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.WantFlipped(t, g, id, "Goka the Unjust", []string{"R"}, 4)
	bear := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Their Bear", "oracle-their-bear", 5, 5))
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("Goka targeted an undamaged creature")
	}
	effectstest.DamageAndSettle(t, g, uuid.Nil, bear, 1)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if effects.OnBattlefield(g, bear) {
		t.Error("4 more damage on a 5/5 that already took 1 should kill it")
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func initiateOfBloodRow() cards.Card {
	return flipRow(initiateOfBloodOracleID, "{3}{R}", []string{"R"}, nil,
		cards.CardFace{Name: "Initiate of Blood", TypeLine: "Creature — Ogre Shaman",
			OracleText: "{T}: This creature deals 1 damage to target creature that was dealt damage this turn. When that creature dies this turn, flip this creature.",
			Power:      "2", Toughness: "2"},
		cards.CardFace{Name: "Goka the Unjust", TypeLine: "Legendary Creature — Ogre Shaman",
			OracleText: "{T}: Goka deals 4 damage to target creature that was dealt damage this turn.",
			Power:      "4", Toughness: "4"})
}

func initiateTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := flipMain(t)
	return g, me, opp, flipEntered(t, g, initiateOfBloodRow(), me)
}

// The Initiate can only aim at a creature that was dealt damage this
// turn.
func TestInitiateOfBloodNeedsADamagedCreature(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 3, 3))
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("targeted a creature that was not dealt damage this turn")
	}
	damageAndSettle(t, g, uuid.Nil, bear, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("a damaged creature is a legal target: %v", err)
	}
}

// The damage kills the creature it was aimed at: "when that creature
// dies this turn" flips the Initiate, as the card intends.
func TestInitiateOfBloodFlipsWhenItsDamageKillsTheCreature(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 2, 2))
	damageAndSettle(t, g, uuid.Nil, bear, 1)
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if onBattlefield(g, bear) {
		t.Fatal("the bear survived 2 damage")
	}
	wantFlipped(t, g, id, "Goka the Unjust", []string{"R"}, 4)
}

// A creature that survives only flips the Initiate if it dies later
// the same turn, and not on a later turn.
func TestInitiateOfBloodFlipsOnlyIfThatCreatureDiesThisTurn(t *testing.T) {
	t.Run("dies later this turn", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 3, 3))
		damageAndSettle(t, g, uuid.Nil, bear, 1)
		p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
		wantUnflipped(t, g, id, "Initiate of Blood")
		damageAndSettle(t, g, uuid.Nil, bear, 1)
		passPriorityAroundTable(t, g)
		wantFlipped(t, g, id, "Goka the Unjust", []string{"R"}, 4)
	})
	t.Run("dies next turn", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 3, 3))
		damageAndSettle(t, g, uuid.Nil, bear, 1)
		p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
		advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
		damageAndSettle(t, g, uuid.Nil, bear, 5)
		passPriorityAroundTable(t, g)
		wantUnflipped(t, g, id, "Initiate of Blood")
	})
	t.Run("a different creature dies", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 3, 3))
		other := apaPush(g, opp.ID, opp.ID, apaCreature("Other", "oracle-other", 1, 1))
		damageAndSettle(t, g, uuid.Nil, bear, 1)
		p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
		damageAndSettle(t, g, uuid.Nil, other, 1)
		passPriorityAroundTable(t, g)
		wantUnflipped(t, g, id, "Initiate of Blood")
	})
	t.Run("the Initiate is gone", func(t *testing.T) {
		g, me, opp, id := initiateTable(t)
		bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 3, 3))
		damageAndSettle(t, g, uuid.Nil, bear, 1)
		p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		damageAndSettle(t, g, uuid.Nil, bear, 5)
		passPriorityAroundTable(t, g)
		if c, ok := g.LookupCardForEffect(id); !ok || c.Flipped || c.ActiveFace != 0 {
			t.Errorf("the dead Initiate is flipped %v face %d", c.Flipped, c.ActiveFace)
		}
	})
}

// A returned Initiate is a new object and the old trigger flips
// nothing (CR 400.7).
func TestInitiateOfBloodDoesNotFlipANewObject(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 3, 3))
	damageAndSettle(t, g, uuid.Nil, bear, 1)
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	g.WithWriteLock(func() {
		if err := (BounceToHand{Target: id}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
		if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	damageAndSettle(t, g, uuid.Nil, bear, 5)
	passPriorityAroundTable(t, g)
	wantUnflipped(t, g, id, "Initiate of Blood")
}

// Goka deals 4 to a creature that was dealt damage this turn.
func TestGokaDealsFourToADamagedCreature(t *testing.T) {
	g, me, opp, id := initiateTable(t)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	wantFlipped(t, g, id, "Goka the Unjust", []string{"R"}, 4)
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Their Bear", "oracle-their-bear", 5, 5))
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("Goka targeted an undamaged creature")
	}
	damageAndSettle(t, g, uuid.Nil, bear, 1)
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if onBattlefield(g, bear) {
		t.Error("4 more damage on a 5/5 that already took 1 should kill it")
	}
}

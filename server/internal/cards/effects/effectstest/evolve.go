package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// evolve_test.go — #1805, CR 702.100 (ADR 0106 §3). Evolve is a
// canonical keyword token the engine turns into one trigger per
// instance (game/evolve.go). These tests hold the rule's edges — what
// triggers it, the CR 603.4 re-check on resolution, last-known
// information, instances, grants and the CR 702.100b "evolves" event —
// plus one per card that lands with it.

const (
	OracleTyranidPrime    = "12a43da3-57aa-4394-8cce-e53ec7e3538e"
	OraclePropagatorDrone = "9485db26-7e03-4a66-beeb-627a3bc6f367"
	OracleRenegadeKrasis  = "7b7e622c-815a-44a0-88ff-62bbe8dd5582"
	OracleWatchfulRadstag = "572afa0f-9e82-4899-b622-a7554625b9f4"
	OracleDinosaurEgg     = "ac4d5a97-6177-4eac-b0f2-cf10531ad879"
	OracleScurryOak       = "eee6c02b-7d6a-4445-ad27-8b03176147d4"
	OracleLonis           = "52d8e492-b7b7-4ad4-8f1f-5c50993e55e0"
	OracleGyreSage        = "e3aa16cf-079b-4737-9ffd-7bfdffef0cb2"
	OracleSimicFluxmage   = "68af451d-82f9-4c78-8bb1-36503e3f2e34"
	OraclePollywog        = "99a3cea9-f4ab-4da3-a085-c09dc93fc8cb"
)

// pushEvolveCreature is a creature with no catalog entry whose printed
// evolve the deck importer stamped — Cloudfin Raptor's case, and the
// proof that a vanilla evolve creature needs no card file.
func PushEvolveCreature(g *game.Game, controller uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	if len(keywords) == 0 {
		keywords = []string{game.KeywordEvolve}
	}
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird Mutant",
		Power: power, Toughness: toughness, Keywords: keywords,
		Owner: controller, Controller: controller,
	})
}

// enterCreature puts a creature into owner's hand and moves it onto the
// battlefield through the real entry path, which harvests its ETB
// triggers onto the stack before it returns.
func EnterCreature(t *testing.T, g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	t.Helper()
	return EnterCard(t, g, owner, game.Card{Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: toughness})
}

func EnterCard(t *testing.T, g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	PlayerByIDForTest(g, owner).Hand.PushTop(c)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		c.InstanceID,
	); err != nil {
		t.Fatalf("enter %s: %v", c.Name, err)
	}
	return c.InstanceID
}

func PlusOnes(g *game.Game, id uuid.UUID) int { return CountersOn(g, id, game.CounterPlusOne) }

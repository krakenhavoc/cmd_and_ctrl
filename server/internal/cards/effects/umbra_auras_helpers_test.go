package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// umbra_auras_helpers_test.go is the setup the umbra armor auras share
// (#3654): a creature to enchant, and the two checks every umbra makes
// the same way. Each card's own test file pins what is particular to it.

// umbraCreature puts a creature with the given printed P/T under owner's
// control. The Aura goes on it through auraCast, so the cast path does
// the targeting.
func umbraCreature(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// assertUmbraSavesFromDestroy destroys the creature and checks the umbra
// took the destruction: the creature stays, the Aura goes to me's
// graveyard instead.
func assertUmbraSavesFromDestroy(t *testing.T, g *game.Game, me *game.Player, bear, umbra uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("umbra armor should save the creature from a destroy")
	}
	if !me.Graveyard.Contains(umbra) {
		t.Fatal("the umbra should be destroyed instead")
	}
}

// assertEffectiveKeyword checks that a keyword is among the creature's
// effective abilities, granted or printed.
func assertEffectiveKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) {
	t.Helper()
	var has bool
	g.ReadSnapshot(func() { has = effectiveHasKeyword(findBattlefieldCardForTest(g, id), kw) })
	if !has {
		t.Errorf("the creature does not have %s", kw)
	}
}

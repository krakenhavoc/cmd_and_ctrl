package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// darksteel_forge_test.go — the grant has to reach the destruction
// path through Effective().Abilities, not through Card.Keywords. A
// guard that read the printed slice would leave every card the Forge
// protects destroyable while the badge said otherwise, and only a
// GRANT test catches that; Darksteel Citadel's own printed keyword
// would pass either way.

const darksteelForgeOracle = "9b3bec05-441f-4fdf-8b51-69fa8613fcd4"

func forgePermanent(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	c := game.NewCard(name, owner)
	c.TypeLine = typeLine
	c.OracleID = oracleID
	c.Controller = owner
	if c.IsCreature() {
		c.Power, c.Toughness = 2, 2
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func forgeStillThere(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

func TestDarksteelForgeProtectsArtifactsIncludingItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID

	forge := forgePermanent(g, me, "Darksteel Forge", "Artifact", darksteelForgeOracle)
	myRock := forgePermanent(g, me, "Mana Rock", "Artifact", "")
	myBear := forgePermanent(g, me, "Grizzly Bears", "Creature — Bear", "")
	theirRock := forgePermanent(g, opp, "Their Rock", "Artifact", "")

	if abilities := effectiveAbilities(t, g, myRock); !containsString(abilities, "indestructible") {
		t.Fatalf("grant missing on your artifact: %v", abilities)
	}
	if abilities := effectiveAbilities(t, g, theirRock); containsString(abilities, "indestructible") {
		t.Fatalf("grant leaked to an opponent's artifact: %v", abilities)
	}

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{forge, myRock, myBear, theirRock} {
			_ = g.DestroyPermanentForEffect(id)
		}
	})

	if !forgeStillThere(g, forge) {
		t.Error("the Forge protects itself — it is an artifact you control")
	}
	if !forgeStillThere(g, myRock) {
		t.Error("your other artifact should have survived")
	}
	if forgeStillThere(g, myBear) {
		t.Error("a non-artifact creature is not covered and must die")
	}
	if forgeStillThere(g, theirRock) {
		t.Error("an opponent's artifact is not covered and must die")
	}
}

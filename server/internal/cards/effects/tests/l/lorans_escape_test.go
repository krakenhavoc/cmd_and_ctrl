package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const loransEscapeOracle = "1a9048d6-f3fa-4f3b-9baa-35318177bc6e"

// TestLoransEscapeGrantsHexproofAndIndestructibleThenScries checks the
// keyword grant survives a Doom Blade / Wrath-shaped removal attempt
// and that the scry still happens.
func TestLoransEscapeGrantsHexproofAndIndestructibleThenScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom Me")
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Loran's Escape", "Instant", loransEscapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "hexproof") || !effectstest.EffectiveAbilitiesContain(t, g, bear, "indestructible") {
		t.Error("the targeted creature should have hexproof and indestructible until end of turn")
	}
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil {
		t.Fatalf("no scry-1 prompt: %+v", g.PendingChoices)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if !g.Battlefield.Contains(bear) {
		t.Error("indestructible should have stopped the destroy")
	}
}

// TestLoransEscapeAcceptsAnArtifactTarget — "target artifact or
// creature".
func TestLoransEscapeAcceptsAnArtifactTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom Me")
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Loran's Escape", "Instant", loransEscapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.EffectiveAbilitiesContain(t, g, rock, "indestructible") {
		t.Error("an artifact target should also get the grant")
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const loransEscapeOracle = "1a9048d6-f3fa-4f3b-9baa-35318177bc6e"

// TestLoransEscapeGrantsHexproofAndIndestructibleThenScries checks the
// keyword grant survives a Doom Blade / Wrath-shaped removal attempt
// and that the scry still happens.
func TestLoransEscapeGrantsHexproofAndIndestructibleThenScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me")
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Loran's Escape", "Instant", loransEscapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !effectiveAbilitiesContain(t, g, bear, "hexproof") || !effectiveAbilitiesContain(t, g, bear, "indestructible") {
		t.Error("the targeted creature should have hexproof and indestructible until end of turn")
	}
	if c := scryChoiceFor(g, me.ID); c == nil {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me")
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Loran's Escape", "Instant", loransEscapeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)

	if !effectiveAbilitiesContain(t, g, rock, "indestructible") {
		t.Error("an artifact target should also get the grant")
	}
}

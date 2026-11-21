package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auras_294a_returns_test.go — the slice 294-a Auras that move cards
// between zones: Kaya's Ghostform, Gift of Immortality, Sheltered by
// Ghosts, Mantle of the Ancients, and Mechanized Production's token.

const (
	WildGrowthAuraOracle = "706ae742-1807-44b7-a4fa-f2e26f61519a"
	PacifismAuraOracle   = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
)

// namedOnBattlefield counts battlefield cards with the name under a
// controller.
func NamedOnBattlefield(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// killCreature moves a battlefield creature to its owner's graveyard
// through the public mutator, which runs the state-based actions on
// the way out — so a fallen-off Aura is swept before any trigger
// goes on the stack, as it is in play.
func KillCreature(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
		t.Fatalf("MoveCardByID to the graveyard: %v", err)
	}
}

// exileCreature is killCreature for exile.
func ExileCreature(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("MoveCardByID to exile: %v", err)
	}
}

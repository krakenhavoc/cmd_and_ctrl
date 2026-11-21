package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// craft_test.go — CR 702.167, craft (#2124, ADR 0137). The seam is
// pinned here with real cards, through the import road, because the
// return is a transform and a flat fixture has no back face (see
// transform_cards_test.go).
//
// What each test pins:
//
//   - the materials are a cost spanning two zones (CR 702.167b): a
//     permanent you control, a card in your own graveyard, or both;
//   - they are validated before anything is paid (CR 118.3), and the
//     graveyard is the ACTIVATOR's, judged on the card's front face
//     there (CR 712.8a);
//   - the card returns on its back face as a new object, under its
//     owner's control (CR 702.167a, CR 400.7);
//   - the new permanent remembers what was exiled to make it
//     (CR 702.167c), and only what is still in exile.

func TithingBladeRow() cards.Card {
	return TransformRow(cardsth.TithingBladeOracleID, "Tithing Blade", "Artifact", "{1}{B}",
		"Consuming Sepulcher", "Artifact", "", "", []string{"B"})
}

// pushCraftCard puts an imported transform card onto the battlefield
// under `p`, without running its enters trigger (the craft is the
// subject, not the front face).
func PushCraftCard(g *game.Game, p *game.Player, row cards.Card) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	return PushBattlefieldCardWithTimestamp(g, c)
}

// craftedPermanent finds the one battlefield permanent showing `name`.
func CraftedPermanent(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	var found []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				found = append(found, c)
			}
		}
	})
	if len(found) != 1 {
		t.Fatalf("%d permanents named %q on the battlefield, want 1", len(found), name)
	}
	return found[0]
}

func ActivateCraft(g *game.Game, p *game.Player, source uuid.UUID, index int, materials ...uuid.UUID) error {
	return g.ActivateCatalogAbility(p.ID, source, index, game.ActivateAbilityParams{
		Strict:            true,
		ExilePermanentIDs: materials,
	})
}

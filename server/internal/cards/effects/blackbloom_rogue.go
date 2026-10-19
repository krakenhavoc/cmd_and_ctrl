package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blackbloom Rogue // Blackbloom Bog — Modal DFC. Front: Creature —
// Human Rogue {2}{B}, 2/3, Menace; "This creature gets +3/+0 as long as
// an opponent has eight or more cards in their graveyard." Back: Land
// (mdfc_lands.go).
//
// A layer 7c modification of itself, read live on every layer pass, so
// the bonus turns on and off as graveyards fill and empty. "An opponent"
// is any seat other than the controller that is still in the game; the
// graveyard is counted per player, never summed. No simplification.
func init() {
	Register(Spec{
		OracleID:        "34320ebf-da97-44a4-bbeb-a9da06548289",
		Name:            "Blackbloom Rogue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +3/+0 as long as an opponent has eight or more cards in their graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					anOpponentHasGraveyardOfAtLeast(g, source.Controller, 8)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 3
			},
		}},
	})
}

// anOpponentHasGraveyardOfAtLeast is "an opponent has n or more cards in
// their graveyard": one opponent's own graveyard, not the sum.
func anOpponentHasGraveyardOfAtLeast(g *game.Game, controller uuid.UUID, n int) bool {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller || p.Graveyard == nil {
			continue
		}
		if len(p.Graveyard.Cards) >= n {
			return true
		}
	}
	return false
}

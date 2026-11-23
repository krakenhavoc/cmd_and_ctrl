package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mutagen Connoisseur — Creature — Vedalken Mutant {1}{G}{U}, 0/5
// (March of the Machine):
//
//	"Flying, vigilance
//	 This creature gets +1/+0 for each transformed permanent you control."
//
// A transformed permanent is a double-faced permanent with its back face up
// (CR 701.27g; TransformedPermanent). The count is read on every layer pass,
// so it moves as permanents transform, enter transformed or leave; the
// Connoisseur is not itself double-faced and never counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "07b4f5b5-47ef-43a5-af71-51693d775d4b",
		Name:            "Mutagen Connoisseur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+0 for each transformed permanent you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += TransformedPermanentsControlledBy(g, source.Controller)
			},
		}},
	})
}

// transformedPermanentsControlledBy counts the transformed permanents a
// player controls.
func TransformedPermanentsControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && TransformedPermanent(c) {
			n++
		}
	}
	return n
}

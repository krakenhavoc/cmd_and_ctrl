package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhox Oracle — Creature — Rhino Monk {4}{G}, 4/2 (Core Set 2019, #5527):
//
//	"When this creature enters, draw a card."
//
// Elvish Visionary's text: draw a card as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d26d1cce-3bcf-48d4-abce-8b12ca7b7432",
		Name:         "Rhox Oracle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rhox Oracle — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

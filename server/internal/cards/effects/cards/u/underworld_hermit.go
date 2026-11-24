package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Hermit — Creature — Human Peasant {4}{B}{B}, 3/3 (Modern Horizons 2):
//
//	"When this creature enters, create a number of 1/1 green Squirrel creature
//	 tokens equal to your devotion to black. (Each {B} in the mana costs of
//	 permanents you control counts toward your devotion to black.)"
//
// Devotion (CR 700.5) is counted as the trigger resolves, so the Hermit's own
// {B}{B} is in the count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7f4ddc9-cea0-42f9-bbca-2ad87a24bfc8",
		Name:         "Underworld Hermit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Underworld Hermit — create a number of 1/1 green Squirrel creature tokens equal to your devotion to black",
				effects.OldB033UnderworldHermitSquirrels),
		},
	})
}

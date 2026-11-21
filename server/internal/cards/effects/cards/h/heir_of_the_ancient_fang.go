package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heir of the Ancient Fang — Creature — Snake Samurai {2}{G}, 2/3 (Kamigawa:
// Neon Dynasty):
//
//	"This creature enters with a +1/+1 counter on it if you control a
//	 modified creature. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// Read as it enters, inside the entry replacement (CR 614.1c), so the Heir
// is not yet on the battlefield and cannot modify itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bcdbded-46f0-44b9-86ad-edd6414e7ef6",
		Name:         "Heir of the Ancient Fang",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				if effects.NeoB10ControlsModified(g, src.Controller) {
					return 1
				}
				return 0
			}, "Heir of the Ancient Fang: enters with a +1/+1 counter if you control a modified creature"),
		},
	})
}

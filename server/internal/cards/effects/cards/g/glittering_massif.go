package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glittering Massif — Land — Mountain Plains (Marvel Super Heroes
// Commander, #4455): "({T}: Add {R} or {W}.) This land enters tapped.
// Cycling {2}". The red-white member of the cycling-land cycle, the
// same shape as Coastal Peak and Canyon Slough.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "23732d20-aba6-46cc-89d0-0542629bc6b5",
		Name:          "Glittering Massif",
		Completeness:  effects.CompletenessFull,
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("R", "W")},
	})
}

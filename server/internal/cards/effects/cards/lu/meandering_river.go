package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meandering River — Land (Oath of the Gatewatch, #5545):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}."
//
// The Guildgates' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "2d9663be-c466-4191-85e2-a69ce0965432",
		Name:          "Meandering River",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("W", "U")},
	})
}

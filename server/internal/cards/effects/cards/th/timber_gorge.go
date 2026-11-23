package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timber Gorge — Land (Oath of the Gatewatch, #5545):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}."
//
// The Guildgates' shape: it enters tapped (a replacement, CR 614.1c) and
// taps for either colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "00b34fab-5a80-4a4d-b6cf-72479197677a",
		Name:          "Timber Gorge",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("R", "G")},
	})
}

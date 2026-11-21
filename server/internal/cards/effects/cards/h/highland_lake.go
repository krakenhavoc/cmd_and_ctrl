package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Highland Lake — Land (Shadows over Innistrad, #5543):
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}."
//
// The plain tapped dual, the Guildgates' shape without the Gate type:
// SelfEntersTapped and dualManaAbility (guildgates.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "c643365a-4255-4d23-adb9-0f8b456f0838",
		Name:          "Highland Lake",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("U", "R")},
	})
}

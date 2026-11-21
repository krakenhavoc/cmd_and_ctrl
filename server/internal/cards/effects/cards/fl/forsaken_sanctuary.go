package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forsaken Sanctuary — Land (Shadows over Innistrad, #5543):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}."
//
// The plain tapped dual, the Guildgates' shape without the Gate type:
// SelfEntersTapped and dualManaAbility (guildgates.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "941b0dd1-0df2-48ee-8829-615e9c3177a7",
		Name:          "Forsaken Sanctuary",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("W", "B")},
	})
}

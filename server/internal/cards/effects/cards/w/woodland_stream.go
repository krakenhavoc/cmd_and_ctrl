package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Woodland Stream — Land (Shadows over Innistrad, #5543):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}."
//
// The plain tapped dual, the Guildgates' shape without the Gate type:
// SelfEntersTapped and dualManaAbility (guildgates.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e887fb3f-d4c9-4022-8f75-1de6ec94af96",
		Name:          "Woodland Stream",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("G", "U")},
	})
}

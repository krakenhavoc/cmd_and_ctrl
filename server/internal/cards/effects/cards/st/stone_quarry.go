package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Quarry — Land (Shadows over Innistrad, #5543):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}."
//
// The plain tapped dual, the Guildgates' shape without the Gate type:
// SelfEntersTapped and dualManaAbility (guildgates.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "90bccf66-58ec-445d-ac96-c6013054a1b4",
		Name:          "Stone Quarry",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("R", "W")},
	})
}

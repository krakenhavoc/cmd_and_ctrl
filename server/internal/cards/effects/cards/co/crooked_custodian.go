package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crooked Custodian (Streets of New Capenna, #5491):
//
//	This creature enters tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c39bc53-52d3-4610-920e-5607a625b01f",
		Name:         "Crooked Custodian",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}

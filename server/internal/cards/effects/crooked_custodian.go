package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crooked Custodian (Streets of New Capenna, #5491):
//
//	This creature enters tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c39bc53-52d3-4610-920e-5607a625b01f",
		Name:         "Crooked Custodian",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
	})
}

package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gohn, Town of Ruin — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}."
//
// Treno, Dark City's shape in black and green: the enters-tapped self
// replacement (CR 614.1d) and the declared two-colour mana ability. The
// Town subtype rides the printed type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "13fbd525-c9ee-4bd6-b269-94fb264024b6",
		Name:          "Gohn, Town of Ruin",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("B", "G")},
	})
}

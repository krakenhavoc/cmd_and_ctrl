package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guadosalam, Farplane Gateway — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}."
//
// Treno, Dark City's shape in green and blue: the enters-tapped self
// replacement (CR 614.1d) and the declared two-colour mana ability. The
// Town subtype rides the printed type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "849c97e4-df15-4ecb-bdf8-283bb497d90c",
		Name:          "Guadosalam, Farplane Gateway",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("G", "U")},
	})
}

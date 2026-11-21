package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gongaga, Reactor Town — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}."
//
// Treno, Dark City's shape in red and green: the enters-tapped self
// replacement (CR 614.1d) and the declared two-colour mana ability. The
// Town subtype rides the printed type line.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ac4e1818-4873-4143-9b99-46f205d006d9",
		Name:          "Gongaga, Reactor Town",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("R", "G")},
	})
}

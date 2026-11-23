package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Insomnia, Crown City — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}."
//
// The white-black tapped Town, Treno's shape: the enters-tapped self
// replacement (CR 614.1d) and the declared two-colour mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b7140f72-91b6-4be4-8d71-c019e64c2351",
		Name:          "Insomnia, Crown City",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("W", "B")},
	})
}

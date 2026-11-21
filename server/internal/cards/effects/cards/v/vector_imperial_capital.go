package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vector, Imperial Capital — Land — Town (Final Fantasy, #4462):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}."
//
// Baron, Airship Kingdom's shape: the enters-tapped self-replacement
// (CR 614.1c, so nobody gets a window to answer it) and the two-colour
// mana ability every dual in the catalog uses. Registered because the
// engine's synthetic mana ability only covers basic lands. Town is a
// land subtype with no rules of its own; the type line carries it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "8dcba63c-4701-4dc9-81f5-ca8e8933a3ba",
		Name:          "Vector, Imperial Capital",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("B", "R")},
	})
}

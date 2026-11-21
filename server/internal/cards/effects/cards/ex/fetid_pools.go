package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fetid Pools — Land — Island Swamp:
//
//	"({T}: Add {U} or {B}.)
//	 This land enters tapped.
//	 Cycling {2}"
//
// The Amonkhet cycling dual. The reminder-text mana ability is
// declared, not derived (the synthetic basic-land ability only fires
// for the BASIC supertype), exactly as the Triomes do; cycling is the
// shared hand ability from cycling.go, and the tapped entry is the
// self-replacement rather than an ETB tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "32b03b48-06da-4a74-a7ac-e5ae39a4f428",
		Name:          "Fetid Pools",
		Completeness:  effects.CompletenessFull,
		Replacements:  []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("U", "B")},
		Activated:     []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}

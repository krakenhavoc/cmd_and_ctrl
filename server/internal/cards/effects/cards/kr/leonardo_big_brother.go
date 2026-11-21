package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonardo, Big Brother — Legendary Creature — Mutant Ninja Turtle
// {2}{W}, 1/3 (TMT, #4457):
//
//	"Sneak {W} (… He enters tapped and attacking.)
//	 Leonardo gets +1/+0 for each other creature you control."
//
// Sneak is Sneak(cost) (CR 702.190). The bonus is a layer 7c static
// (PumpSelfCreatureOrAttachedPer, which on a creature that is not an
// Aura applies to itself alone), counted on every pass.
//
// No simplification.
func init() {
	pump := effects.PumpSelfCreatureOrAttachedPer(1, 0, effects.TmtB11OtherCreaturesYouControl)
	pump.Label = "Leonardo gets +1/+0 for each other creature you control."
	effects.Register(effects.Spec{
		OracleID:         "6496eea7-38f5-4428-a379-b08b0eeb5239",
		Name:             "Leonardo, Big Brother",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{W}")},
		Static:           []game.StaticAbility{pump},
	})
}

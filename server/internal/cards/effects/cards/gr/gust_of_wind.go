package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gust of Wind — Sorcery {3}{U} (Foundations Jumpstart):
//
//	"This spell costs {2} less to cast if you control a creature with
//	 flying.
//	 Return target nonland permanent you don't control to its owner's hand.
//	 Draw a card."
//
// Draconic Lore's self cost modifier over a creature with flying, then
// Leave in the Dust's bounce-and-draw over a nonland permanent an
// opponent controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cfc14ce-9940-4c61-9daa-2c6dbf1a80ad",
		Name:         "Gust of Wind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you control a creature with flying.",
				effects.YouControlA(effects.And(effects.Creature(), effects.HasKeyword("flying")))),
		},
		Targets:   effects.TargetPermanent("target nonland permanent you don't control", effects.And(effects.Nonland(), effects.OpponentControls())),
		OnResolve: effects.BounceTargetThenDrawACard,
	})
}

package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bellowing Saddlebrute — Creature — Orc Warrior {3}{B}, 4/5 (Khans of Tarkir):
//
//	"Raid — When this creature enters, you lose 4 life unless you attacked
//	 this turn."
//
// Not an intervening "if": the trigger always goes on the stack and the
// attack record is read as it resolves (CR 608.2c). The life loss is not
// damage (CR 119.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1e5ad8c-d425-44ad-a8f6-e9290667f6f5",
		Name:         "Bellowing Saddlebrute",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bellowing Saddlebrute — raid: you lose 4 life unless you attacked this turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if effects.B18AttackedThisTurn(ctx.Game, ctx.Controller()) {
						return nil
					}
					return effects.YouLoseLife(4)(ctx)
				}))),
		},
	})
}

package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caligo Skin-Witch — Creature — Human Wizard {1}{B}, 1/3 (Dominaria, #5530):
//
//	"Kicker {3}{B} (You may pay an additional {3}{B} as you cast this spell.)
//	 When this creature enters, if it was kicked, each opponent discards two cards."
//
// A kicked enters trigger (CR 603.4): each opponent discards two cards
// of their choice, one prompt each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e62c943e-03ac-4a8a-ad3d-faa76b79824c",
		Name:          "Caligo Skin-Witch",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Caligo Skin-Witch — kicked: each opponent discards two cards",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					effects.EachOpponentDiscards(ctx, 2)
					return nil
				}))),
		},
	})
}

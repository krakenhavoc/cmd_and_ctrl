package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anje's Ravager — Creature — Vampire Berserker {2}{R}, 3/3:
//
//	"This creature attacks each combat if able.
//	 Whenever this creature attacks, discard your hand, then draw three
//	 cards.
//	 Madness {1}{R}"
//
// Alexios's AttacksEachCombat static and the engine's madness keyword
// (Basking Rootwalla's seam). The attack trigger is the wheel of Ajani
// Unrelenting's -2: the whole hand is discarded, and the draw waits for a
// prompt a discard may ask (a replacement offer; CR 608.2c) before it
// draws three, so a hand discarded to the Ravager's own trigger can
// madness it back in.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f1016c7-0551-4813-987d-36abf3b2ffeb",
		Name:         "Anje's Ravager",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{R}",
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Anje's Ravager — discard your hand, then draw three cards",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
						return err
					}),
					effects.DrawCards{N: 3},
				)),
		},
	})
}

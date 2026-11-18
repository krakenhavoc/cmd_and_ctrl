package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8f1016c7-0551-4813-987d-36abf3b2ffeb",
		Name:         "Anje's Ravager",
		Completeness: CompletenessFull,
		Madness:      "{1}{R}",
		Static:       []game.StaticAbility{AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Anje's Ravager — discard your hand, then draw three cards",
				Do(
					Instruction(func(ctx *Context) error {
						_, err := discardWholeHand(ctx.Game, ctx.Controller())
						return err
					}),
					DrawCards{N: 3},
				)),
		},
	})
}

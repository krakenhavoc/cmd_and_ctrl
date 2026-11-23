package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divide by Zero — Instant {2}{U} (EDHREC rank 17095):
//
//	"Return target spell or permanent with mana value 1 or greater to
//	 its owner's hand.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// One target that is either a spell on the stack or a permanent, with
// the mana value clause on both halves (X is zero on a permanent and the
// chosen value on a spell, CR 202.3e; a token with no cost is zero and
// is not a legal target). The return is Venser, Shaper Savant's, shared
// as ReturnTargetSpellOrPermanentToHand; learn follows it as a step
// (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b1c1a0c-f443-4eb5-a430-9eea886361ca",
		Name:         "Divide by Zero",
		Purpose:      game.Purpose{Learn: 1},
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetSpellOrPermanent("target spell or permanent with mana value 1 or greater",
			effects.ManaValueGE(1), effects.ManaValueGE(1)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnTargetSpellOrPermanentToHand(ctx.Game, ctx.Item)
			}),
			effects.Learn{},
		),
	})
}

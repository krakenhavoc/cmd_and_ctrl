package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whirlpool Whelm — Instant {1}{U} (Commander 2011, #5574):
//
//	"Clash with an opponent, then return target creature to its owner's hand.
//	 If you win, you may put that creature on top of its owner's library
//	 instead."
//
// The clash comes first and the creature moves after it, as printed: its
// target is checked again when the clash is over (CR 608.2b), and a creature
// that has gone does nothing. Winning asks the controller whether to put it on
// top of its owner's library instead; losing or a tie returns it to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f527852-bf7d-43a3-be0b-784c73ed88aa",
		Name:         "Whirlpool Whelm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ClashWithAnOpponent{Then: ctx.ItemTail(whirlpoolWhelmClashed)}.Apply(ctx)
		},
	})
}

// whirlpoolWhelmClashed is "then return target creature to its owner's hand.
// If you win, you may put it on top of its owner's library instead". An
// on-disk key.
var whirlpoolWhelmClashed = effects.CardTail("then/whirlpool-whelm-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
		return nil
	}
	if !effects.WonTheClash(o) {
		return effects.BounceTheTarget(ctx.Item, ctx)
	}
	return effects.MayChoice{
		Question: "Whirlpool Whelm — put that creature on top of its owner's library instead of returning it to hand?",
		YesLabel: "Put it on top of the library",
		NoLabel:  "Return it to hand",
		Then:     whirlpoolWhelmTopOrHand,
	}.Apply(ctx)
})

// whirlpoolWhelmTopOrHand is the answer: yes tucks the creature on top of its
// owner's library, no returns it to hand. An on-disk key.
var whirlpoolWhelmTopOrHand = effects.MayYesNo("confirm/whirlpool-whelm-top-or-hand",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.PutChosenTargetOnTopOfLibrary(ctx.Game, ctx.Item)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error { return effects.BounceTheTarget(ctx.Item, ctx) })

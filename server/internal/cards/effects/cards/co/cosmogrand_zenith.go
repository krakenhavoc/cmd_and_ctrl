package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cosmogrand Zenith — Creature — Human Soldier {2}{W}, 2/4 (Edge of
// Eternities, #4461):
//
//	"Whenever you cast your second spell each turn, choose one —
//	 • Create two 1/1 white Human Soldier creature tokens.
//	 • Put a +1/+1 counter on each creature you control."
//
// The trigger is Alphinaud Leveilleur's YouCastYourSecondSpellEachTurn (the
// cast path bumps the per-turn tally before it emits EventCast, so a total
// of exactly two is the second spell); a spell cast before the Zenith was on
// the battlefield still counts toward the two. The mode is chosen as the
// trigger goes on the stack (CR 603.3c), through the mode_pick prompt. The
// counter bullet reads the creatures as it resolves, so the spell that
// triggered it, if it is a creature spell, is still on the stack and gets
// none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2cd65f6-7cbc-497c-bdc9-698ff17da31e",
		Name:         "Cosmogrand Zenith",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{cosmograndZenithTrigger()},
	})
}

// cosmograndZenithTrigger is the second-spell trigger with its two bullets.
func cosmograndZenithTrigger() game.TriggeredAbility {
	t := effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
		"Cosmogrand Zenith — you cast your second spell this turn",
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Create two 1/1 white Human Soldier creature tokens.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 2}.Apply(ctx)
			}),
		effects.ModeDoing("Put a +1/+1 counter on each creature you control.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.PutACounterOnEachCreatureYouControl(game.CounterPlusOne)(ctx.Game, item)
			}),
	)
	return t
}

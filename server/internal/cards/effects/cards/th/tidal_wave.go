package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tidal Wave — Instant {2}{U} (Eternal Masters, S82 OLD, #5542):
//
//	"Create a 5/5 blue Wall creature token with defender. Sacrifice it at the
//	 beginning of the next end step."
//
// The token is made, then a delayed trigger (CR 603.7) sacrifices that object
// at the next end step (Blastoneuron's shape, without the haste): cast in your
// own end step it waits for the next turn's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5f5b7e5-af1e-4c8a-8d44-3e2152955d4e",
		Name:         "Tidal Wave",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     item.SourceCardID,
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("5/5 blue Wall with defender")), Count: 1}},
			}, ctx.ItemTail(tidalWaveSacrifice))
		},
	})
}

// tidalWaveSacrifice schedules the sacrifice of the Wall at the next end step.
// An on-disk key: never renamed, never reused.
var tidalWaveSacrifice = effects.CardTail("then/s82-old-b279-tidal-wave-sacrifice", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Tidal Wave — sacrifice the Wall",
		Cards:      o.IDs,
		Body:       effects.SacrificeListedCardsBody,
	}.Apply(ctx)
})

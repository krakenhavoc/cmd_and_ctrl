package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waylay — Instant {2}{W} (Urza's Saga, S82 OLD batch 729, #5624):
//
//	"Create three 2/2 white Knight creature tokens. Exile them at the
//	 beginning of the next cleanup step."
//
// The tokens that actually landed are handed to a CR 603.7 delayed trigger
// that fires as the next cleanup step begins (CR 514.3a), so a token a
// replacement kept off the battlefield is simply not exiled, and cast during
// that cleanup step's own priority window the exile waits for the next one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0db78c7-09b5-4981-8ca8-6dbab597c5d9",
		Name:         "Waylay",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("2/2 white Knight")), Count: 3}},
			}, ctx.ItemTail(waylayExile))
		},
	})
}

// waylayExile is "exile them at the beginning of the next cleanup step". An
// on-disk key.
var waylayExile = effects.CardTail("waylay/exile-the-knights", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepCleanup,
		Controller: ctx.Controller(),
		Label:      "Waylay — exile the Knights",
		Cards:      append([]uuid.UUID(nil), o.IDs...),
		Body:       effects.ExileListedCardsBody,
	}.Apply(ctx)
})

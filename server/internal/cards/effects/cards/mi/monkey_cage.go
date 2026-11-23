package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monkey Cage — Artifact {5} (Jumpstart 2022, #5948):
//
//	"When a creature enters, sacrifice this artifact and create X 2/2 green
//	 Monkey creature tokens, where X is that creature's mana value."
//
// "A creature" is any creature, yours or not. X is the entering creature's
// mana value (last-known information, CR 608.2h); the tokens are made
// whether or not the sacrifice happened, since the text has no "if you do".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33a11c01-89a6-446b-84da-e48fd1f27446",
		Name:         "Monkey Cage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33AnyCreatureEntered(ev, g)
			}, "Monkey Cage — sacrifice it and create X 2/2 green Monkey creature tokens, where X is that creature's mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := 0
					if o := item.Trigger.Object; o != nil {
						x = o.ManaValue
					}
					tail := effects.WithN(ctx.Tail(monkeyCageTokens), x)
					if !effects.OnBattlefield(g, ctx.Source()) {
						return monkeyCageMakeMonkeys(ctx, x)
					}
					return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: tail}.Apply(ctx)
				}),
		},
	})
}

// monkeyCageTokens is the "and create X Monkeys" half, run once the sacrifice
// has settled: the text has no "if you do", so it runs whatever the sacrifice
// did. X rides the tail as its N. An on-disk key: never renamed, never reused.
var monkeyCageTokens = effects.CardTail("s82-gap-b144/monkey-cage-tokens", func(ctx *effects.Context, o game.TailOutcome) error {
	return monkeyCageMakeMonkeys(ctx, o.Tail.N)
})

func monkeyCageMakeMonkeys(ctx *effects.Context, x int) error {
	if x <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 green Monkey"), N: x}.Apply(ctx)
}

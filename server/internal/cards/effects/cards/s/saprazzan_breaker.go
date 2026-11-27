package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saprazzan Breaker — Creature — Beast {4}{U}, 3/3 (Mercadian Masques, S82 OLD
// b682, #5619):
//
//	"{U}: Mill a card. If a land card was milled this way, this creature can't
//	 be blocked this turn."
//
// Loafing Giant's mill with a keyed tail that reads what landed (CR 701.17a):
// a land card among the milled makes the Breaker unblockable until end of
// turn, if it is still the same permanent (CR 400.7). An empty library mills
// nothing and so grants nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3413f57-229e-49dd-a776-6905b80b2b12",
		Name:         "Saprazzan Breaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Mill a card. If a land card was milled this way, this creature can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{Player: item.Controller, N: 1, ThenTail: ctx.ItemTail(saprazzanBreakerMilled)}.Apply(ctx)
			},
		}},
	})
}

// saprazzanBreakerMilled is the clause after the mill. An on-disk key: never
// renamed, never reused.
var saprazzanBreakerMilled = effects.CardTail("saprazzan-breaker/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
			return effects.RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked,
				Label: "Saprazzan Breaker — can't be blocked this turn"}.Apply(ctx)
		}
	}
	return nil
})

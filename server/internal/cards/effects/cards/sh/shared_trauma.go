package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shared Trauma — Sorcery {B} (Commander 2011, #5574):
//
//	"Join forces — Starting with you, each player may pay any amount of mana.
//	 Each player mills X cards, where X is the total amount of mana paid this
//	 way."
//
// The join forces process (S82OldB456JoinForces): each living seat in turn
// order from the caster pays any amount of its own mana, then every player
// mills the total, in the same order. A total of zero mills nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef9b7588-9bf0-4632-ae5d-1dfebc4d88eb",
		Name:         "Shared Trauma",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return sharedTraumaJoin.Start(ctx)
		},
	})
}

// sharedTraumaJoin is the process. An on-disk key.
var sharedTraumaJoin = effects.NewS82OldB456JoinForces("pay-unless/old-b456-shared-trauma", "Shared Trauma",
	"each player mills X cards",
	func(ctx *effects.Context, total int) error {
		for _, id := range effects.SeatsFromController(ctx.Game, ctx.Controller()) {
			if err := (effects.MillCards{Player: id, N: total}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

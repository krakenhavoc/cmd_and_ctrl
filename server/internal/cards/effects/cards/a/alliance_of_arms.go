package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alliance of Arms — Sorcery {W} (Commander 2011, #5574):
//
//	"Join forces — Starting with you, each player may pay any amount of mana.
//	 Each player creates X 1/1 white Soldier creature tokens, where X is the
//	 total amount of mana paid this way."
//
// The join forces process (S82OldB456JoinForces): each living seat in turn
// order from the caster pays any amount of its own mana, then every player
// creates the total in Soldiers, in the same order. A total of zero creates
// none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be2a4595-4378-4af8-8368-e3b2bb3cb5bb",
		Name:         "Alliance of Arms",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return allianceOfArmsJoin.Start(ctx)
		},
	})
}

// allianceOfArmsJoin is the process. An on-disk key.
var allianceOfArmsJoin = effects.NewS82OldB456JoinForces("pay-unless/old-b456-alliance-of-arms", "Alliance of Arms",
	"each player creates X 1/1 white Soldier creature tokens",
	func(ctx *effects.Context, total int) error {
		for _, id := range effects.SeatsFromController(ctx.Game, ctx.Controller()) {
			if err := (effects.CreateToken{Controller: id, Template: effects.TokenCard("1/1 white Soldier"), N: total}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

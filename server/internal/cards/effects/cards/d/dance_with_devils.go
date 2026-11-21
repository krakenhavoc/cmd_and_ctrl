package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dance with Devils — Instant {3}{R} (Jumpstart 2022, #5948):
//
//	"Create two 1/1 red Devil creature tokens. They have "When this token
//	 dies, it deals 1 damage to any target.""
//
// Spiked Corridor's Devil, the catalog's token template that carries its own
// dies trigger (ADR 0083), twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caeb508a-c1fd-45b5-a14e-e0ac7321c9a9",
		Name:         "Dance with Devils",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 2, CreatureTokens: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.DevilToken(), N: 2}.Apply(ctx)
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spore Swarm — Instant {3}{G} (Foundations Jumpstart):
//
//	"Create three 1/1 green Saproling creature tokens."
//
// Three Saprolings for the caster at instant speed, the shared
// "1/1 green Saproling" token row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8fb599a7-2c61-446b-a7a7-cd5367f9b796",
		Name:         "Spore Swarm",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3, Answers: game.AnswerMakesBlocker},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 green Saproling"),
				N:          3,
			}.Apply(ctx)
		},
	})
}

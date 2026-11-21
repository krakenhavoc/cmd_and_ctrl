package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eliminate — Instant {1}{B} (Core Set 2021, #5511):
//
//	"Destroy target creature or planeswalker with mana value 3 or less."
//
// Long Goodbye without its "can't be countered": a clause of creature or
// planeswalker And ManaValueLE(3), judged at announce and again as it
// resolves. A face-down creature has mana value 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf307f81-488c-4df5-8b08-ba46e3e7f3d9",
		Name:         "Eliminate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target creature or planeswalker with mana value 3 or less",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}

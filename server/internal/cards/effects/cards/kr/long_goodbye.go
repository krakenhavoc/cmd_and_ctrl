package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Long Goodbye — Instant {1}{B} (Murders at Karlov Manor, #4469):
//
//	"This spell can't be countered. (This includes by the ward ability.)
//	 Destroy target creature or planeswalker with mana value 3 or less."
//
// Abrupt Decay's shape: Spec.CantBeCountered, which a ward trigger's
// counter respects like any other, and a target clause of creature or
// planeswalker And ManaValueLE(3). A face-down creature has mana value 0
// and is a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d63ba8a-a5cd-4fdb-bba3-95612be0077c",
		Name:            "Long Goodbye",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target creature or planeswalker with mana value 3 or less",
			effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}

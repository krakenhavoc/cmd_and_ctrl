package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea God's Scorn — Sorcery {4}{U}{U} (Theros Beyond Death, #5514):
//
//	"Return up to three target creatures and/or enchantments to their owners'
//	 hands."
//
// All the still-legal targets return at once (CR 608.2b), each to its owner's
// hand; a token ceases to exist there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2497a5e-ff2b-4024-9b14-634304d583fe",
		Name:         "Sea God's Scorn",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("up to three target creatures and/or enchantments", effects.Or(effects.Creature(), effects.Enchantment())).WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
	})
}

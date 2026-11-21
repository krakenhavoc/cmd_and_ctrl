package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Echoing Truth — Instant {1}{U} (Commander deck reprint):
//
//	"Return target nonland permanent and all other permanents with the
//	 same name as that permanent to their owners' hands."
//
// The target is a real target (CR 115.1); the rest are not, so hexproof
// on a copy changes nothing. The set is the target plus every other
// permanent on the battlefield sharing its post-layer name, read as the
// spell resolves, and they return as one simultaneous event
// (ReturnAllToHand). If the target is illegal the spell does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef100544-3e6d-4661-a1d1-8b02e58c059c",
		Name:         "Echoing Truth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.ReturnAllToHand{Cards: effects.S82RprB17SameNameAsTarget(ctx.Game, t.ID)}.Apply(ctx)
			}
			return nil
		},
	})
}

package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unsubstantiate — Instant {1}{U} (Jumpstart):
//
//	"Return target spell or creature to its owner's hand."
//
// Venser, Shaper Savant's body narrowed to creatures on the battlefield:
// a spell is returned from the stack without being countered, so a spell
// that can't be countered still goes, and a creature is bounced. A
// noncreature permanent is never a legal pick. The target is judged again
// on resolution (CR 608.2b), so one gone in response leaves nothing to do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c387f46f-ad91-4964-8dc6-acefe537fe3c",
		Name:         "Unsubstantiate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpellOrPermanent("target spell or creature", nil, effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnTargetSpellOrPermanentToHand(ctx.Game, item)
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slip On the Ring — Instant {1}{W}:
//
//	"Exile target creature you own, then return it to the battlefield
//	 under your control. The Ring tempts you."
//
// A creature you own but don't control may be the target (2023-06-16
// ruling). It returns as a new object (CR 400.7), and the tempt runs
// once it has: it may be chosen as the Ring-bearer. A token that is
// exiled does not return (CR 111.8), and the Ring still tempts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6b28298-9624-49d7-ad73-fcd1fd3556d2",
		Name:         "Slip On the Ring",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you own", YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			tempt := ringTemptsYouNext(item)
			ids := legalTargetIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			target := ids[0]
			return ExileTarget{
				Target: target,
				Then: func(ctx *Context, exiled bool) error {
					if !exiled {
						return tempt(ctx.Game)
					}
					return ReturnFromExile{
						Target:     target,
						Controller: ctx.Controller(),
						Then:       func(g *game.Game, _ uuid.UUID) error { return tempt(g) },
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}

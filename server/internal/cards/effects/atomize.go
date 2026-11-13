package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atomize — Instant {2}{B}{G} (Fallout, #5469):
//
//	"Destroy target nonland permanent. Proliferate. (Choose any number
//	 of permanents and/or players, then give each another counter of
//	 each kind already there.)"
//
// Mortality Spear's target and destroy, then the player's proliferate
// choice (Proliferate, #2525), the shape of Hour of Defeat's "destroy,
// then surveil": an indestructible target survives and the proliferate
// still happens. With its only target gone the spell does not resolve
// at all (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b457bc2-2d8c-42cb-9a63-bb4c459d14e3",
		Name:         "Atomize",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return Proliferate{}.Apply(ctx)
		},
	})
}

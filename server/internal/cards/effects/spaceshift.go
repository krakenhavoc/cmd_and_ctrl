package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spaceshift — {1}{U} Instant (Marvel Super Heroes Commander, #4455):
//
//	"Exile target artifact or creature, then return that card to the
//	 battlefield under its owner's control with a +1/+1 counter on it."
//
// A flicker whose return carries the counter (Flicker.PlusOneCounters): the
// counter is part of the entry, so it is there before an enters trigger
// looks (CR 614.1c). The permanent is a new object (CR 400.7). A token
// target is exiled and ceases to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c63a244-75d7-4372-baf4-7a3526f240c8",
		Name:         "Spaceshift",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return Flicker{Target: item.Targets[0].ID, PlusOneCounters: 1}.Apply(ctx)
		},
	})
}

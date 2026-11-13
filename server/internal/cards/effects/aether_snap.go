package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Snap — Sorcery {3}{B}{B} (Darksteel, #5605):
//
//	"Remove all counters from all permanents and exile all tokens."
//
// Two steps in printed order. Every counter of every kind comes off every
// permanent first (removeEveryCounterFrom, the Enchanted River's Grasp
// body, once per permanent), then every token on the battlefield is
// exiled in one sweep. A creature whose toughness came from +1/+1
// counters dies to the state-based check after the spell finishes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a1a16e5-ce4f-42b8-a1c6-05dd635d83cc",
		Name:         "Aether Snap",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{
			Matches: game.SweepNonlandPermanents, How: game.SweepExile, Partial: true,
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if err := removeEveryCounterFrom(ctx, c.InstanceID); err != nil {
					return err
				}
			}
			return ExileAllMatching{Match: IsTokenPredicate()}.Apply(ctx)
		},
	})
}

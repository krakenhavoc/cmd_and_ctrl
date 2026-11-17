package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fertilid's Favor — Instant {3}{G} (March of the Machine):
//
//	"Target player searches their library for a basic land card, puts it
//	 onto the battlefield tapped, then shuffles. Put two +1/+1 counters on up
//	 to one target artifact or creature."
//
// Two target clauses, read by slot (CR 608.2b): the player is Fertilid's
// searcher, who may be an opponent, and the second is up to one artifact or
// creature, which may be none. Each half resolves on its own when the
// other's target is illegal; the spell is countered only when every target
// is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2da8c3e-6ffc-41dd-a68a-62e64f7e9c60",
		Name:         "Fertilid's Favor",
		Completeness: CompletenessFull,
		Targets: TargetPlayer("target player").Then(
			TargetPermanent("up to one target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if p, ok := ctx.ClauseTarget(0); ok {
				if err := (SearchLibrary{Player: p.ID, Predicate: IsBasicLand, Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true,
					TappedOnEntry: true, Reason: "Fertilid's Favor — a basic land, tapped"}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}

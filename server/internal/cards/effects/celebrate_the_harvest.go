package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celebrate the Harvest — Sorcery {3}{G} (Midnight Hunt Commander, S82 MID
// batch 13):
//
//	"Search your library for up to X basic land cards, where X is the number
//	 of different powers among creatures you control. Put those cards onto
//	 the battlefield tapped, then shuffle."
//
// X is read as the spell resolves, through the layers, over the creatures
// its controller controls (rfCreatureCDifferentPowers). A search for basic
// land cards has a stated quality, so it may find fewer (CR 701.23b); the
// library is shuffled even when X is 0 (CR 701.24d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de3764b9-9572-4f5c-8c90-82dedec47557",
		Name:         "Celebrate the Harvest",
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := rfCreatureCDifferentPowers(ctx.Game, ctx.Controller(), ctx.Source())
			if x <= 0 {
				return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
			}
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				TappedOnEntry: true,
				Limit:         x,
				Shuffle:       true,
				Reason:        "Celebrate the Harvest — up to X basic land cards, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}

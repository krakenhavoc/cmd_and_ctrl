package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kodama's Reach — Sorcery — Arcane {2}{G}:
//
//	"Search your library for up to two basic land cards, reveal those
//	cards, put one onto the battlefield tapped and the other into
//	your hand, then shuffle."
//
// Functionally identical to Cultivate (the Arcane subtype matters
// only to splice, which the engine has no notion of), so the
// composition is the same: one search per destination, the first
// deferring its shuffle so the second still sees the library, the
// second shuffling once at the end as the rules require.
//
// The body is duplicated from cultivate.go rather than shared. That
// is deliberate — two files in this package colliding on one helper
// is the exact failure mode the per-card-file convention exists to
// avoid, and the duplication is eight lines.
//
// S22: the second search is CHAINED off the first via Then. The
// searcher now picks, so the first search can still have a prompt
// open when it returns, and a second prompt opened on the next line
// would offer a card that is mid-flight.
func init() {
	Register(Spec{
		OracleID: "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90",
		Name:     "Kodama's Reach",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			source := ctx.Source()
			return SearchLibrary{
				Player:        controller,
				Source:        source,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       false,
				TappedOnEntry: true,
				Reason:        "Kodama's Reach — basic land onto the battlefield tapped",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player:  controller,
						Source:  source,
						Pred:    IsBasicLand,
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  "Kodama's Reach — basic land into your hand",
					})
				},
			}.Apply(ctx)
		},
	})
}

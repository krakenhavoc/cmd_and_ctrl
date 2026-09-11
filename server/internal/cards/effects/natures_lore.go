package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nature's Lore — Sorcery {1}{G}:
//
//	"Search your library for a Forest card, put it onto the
//	battlefield, then shuffle."
//
// UNTAPPED is the entire point — this and Three Visits are premium
// over Rampant Growth because the land is usable the turn it lands,
// so TappedOnEntry is deliberately false here.
//
// "A Forest card" means any land with the Forest subtype, so
// IsLandWithSubtype rather than IsBasicLand: a Snow-Covered Forest
// or a Bayou both qualify in paper and both qualify here.
func init() {
	Register(Spec{
		OracleID: "78826359-fe63-44ad-adc4-a17ffcd710e4",
		Name:     "Nature's Lore",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsLandWithSubtype("forest"),
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Nature's Lore — a Forest card",
			}.Apply(ctx)
		},
	})
}

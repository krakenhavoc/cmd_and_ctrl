package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Explosive Vegetation — Sorcery {3}{G}:
//
//	"Search your library for up to two basic land cards, put them
//	onto the battlefield tapped, then shuffle."
//
// Skyshroud Claim's colour-blind cousin: basics only and tapped, but
// it works in a deck with no duals.
//
// S22: one search with Limit 2. The old two-pass shape was a
// workaround for the first-match picker — with a chooser, the player
// is shown their basics and takes up to two, which is both simpler
// and what the card says.
func init() {
	Register(Spec{
		OracleID:     "a0abd957-01a0-4aa9-8fc8-0e6840d21606",
		Name:         "Explosive Vegetation",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Explosive Vegetation — up to two basic lands",
			}.Apply(ctx)
		},
	})
}

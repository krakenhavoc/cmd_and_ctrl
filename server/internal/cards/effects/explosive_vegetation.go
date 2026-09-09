package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Explosive Vegetation — Sorcery {3}{G}:
//
//	"Search your library for up to two basic land cards, put them
//	onto the battlefield tapped, then shuffle."
//
// Skyshroud Claim's colour-blind cousin: basics only and tapped, but
// it works in a deck with no duals. Same two-pass search, both halves
// to the battlefield rather than one to hand as Cultivate does.
func init() {
	Register(Spec{
		OracleID: "a0abd957-01a0-4aa9-8fc8-0e6840d21606",
		Name:     "Explosive Vegetation",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			if err := (SearchLibrary{
				Player:        controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       false,
				TappedOnEntry: true,
			}).Apply(ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:        controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
			}.Apply(ctx)
		},
	})
}

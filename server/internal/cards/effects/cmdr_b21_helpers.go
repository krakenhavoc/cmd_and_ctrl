package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr_b21_helpers.go — helpers for the Commander deck cards, batch 21
// (#5368). Every name carries the cmdrB21 prefix.

// cmdrB21UpToTwoBasicsTapped is "Search your library for up to two basic
// land cards, put them onto the battlefield tapped, then shuffle."
func cmdrB21UpToTwoBasicsTapped(ctx *Context, name string) error {
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        name + " — up to two basic lands",
	}.Apply(ctx)
}

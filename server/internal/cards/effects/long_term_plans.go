package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Long-Term Plans — Instant {2}{U}:
//
//	"Search your library for a card, then shuffle and put that card
//	 third from the top."
//
// The slow tutor: the card is found, but it arrives three draws from
// now rather than next turn, which is the price of any card for three
// mana. It is Vampiric Tutor's shape — SearchLibrary with ToTop, the
// card placed AFTER the shuffle so it is a known card on an unknown
// library — with the position one field further: Depth 3 (ADR 0088
// Decision 4). A library of fewer than three cards after the search
// takes it on the bottom, which is as close to third from the top as
// that library has.
//
// No reveal: the card does not say "reveal it", so only the caster
// knows what is third from the top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1dfb359-d8f9-491e-88f7-95193b220200",
		Name:         "Long-Term Plans",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneLibrary,
				ToTop:   true,
				Depth:   3,
				Limit:   1,
				Shuffle: true,
				Reason:  "Long-Term Plans — search your library for a card",
			}.Apply(ctx)
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chord of Calling — Instant {X}{G}{G}{G} (EDHREC rank 521):
//
//	"Convoke
//	 Search your library for a creature card with mana value X or
//	 less, put it onto the battlefield, then shuffle."
//
// The instant-speed creature tutor that a wide board pays for by
// tapping itself. Convoke is the S22 tap cost; the engine measures
// its budget against Generic + XSlots×X + the coloured pips, so the
// tapped creatures can pay for X as printed. X is announced at cast
// (ctx.X()) and bounds the search predicate; the searcher picks
// through the S22 search chooser and may fail to find.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "6789a170-f2c5-4fc0-8a45-2b2361e67410",
		Name:     "Chord of Calling",
		TapCost:  Convoke(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsCreature() && manaValueOf(c) <= x },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Chord of Calling — a creature card with mana value X or less",
			}.Apply(ctx)
		},
	})
}

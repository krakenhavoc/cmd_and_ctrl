package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galvanic Iteration — Instant {U}{R}:
//
//	"When you next cast an instant or sorcery spell this turn, copy
//	 that spell. You may choose new targets for the copy."
//	"Flashback {1}{U}{R}"
//
// Doublecast at instant speed with a second use in the graveyard, and
// the reason the flashback matters here rather than being a rider:
// casting it twice in one turn leaves TWO event-conditioned delayed
// triggers owed, and both fire on the next matching cast. Two copies,
// which is the printed outcome and the whole point of the card in a
// storm deck — CR 603.7b makes each trigger fire once, not once
// between them.
//
// Flashback rides the S29 machinery, which binds the two fields that
// must agree: the cost is claimable only out of the graveyard, and
// `CastableZones` has to open that zone or `Register` panics. The
// exile-on-leaving-stack half is what stops it flashing back forever.
func init() {
	Register(Spec{
		OracleID:         "c51ed6e3-813b-49c4-b1de-7f92a77f8f6e",
		Name:             "Galvanic Iteration",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{U}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return WhenYouNextCast(
				"Galvanic Iteration — copy that spell",
				Or(Instant(), Sorcery()),
				copyTheSpellYouJustCast,
			).Apply(ctx)
		},
	})
}

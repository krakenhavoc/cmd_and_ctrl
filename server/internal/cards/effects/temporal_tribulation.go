package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temporal Tribulation — {2}{U} Instant (Star Trek Commander, #4974):
//
//	"Exile two target permanents you control. Return those cards to the
//	 battlefield under their owner's control at the beginning of the next
//	 end step.
//	 Flashback {5}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Eerie Interlude's exile-then-return with exactly two targets: every
// still-legal target is exiled as one event and the cards that reached
// exile come back together at the next end step as new objects
// (CR 400.7). A target that became illegal in response stays where it
// is (CR 608.2b); a token that was exiled never comes back (CR 111.8).
// Flashback is CR 702.34.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "912d6e44-b479-4e07-88a8-9ab5132163f4",
		Name:             "Temporal Tribulation",
		Completeness:     CompletenessFull,
		Targets:          TargetPermanent("two target permanents you control", YouControl()).WithCount(2, 2),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTargetsThenScheduleReturn(ctx, "Temporal Tribulation — return the exiled cards")
		},
	})
}

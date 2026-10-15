package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Spray — Instant {2}{U}:
//
//	"Change the text of target spell or permanent by replacing all
//	 instances of one color word with another or one basic land type
//	 with another until end of turn.
//	 Draw a card."
//
// Mind Bend's either-class change, on a spell or a permanent, for the
// turn only: the record's duration is "until end of turn" (CR 514.2),
// and a change made to a permanent spell goes with the permanent for
// the rest of the turn. The draw is the continuation after the words
// are chosen, so the card is drawn after the change, in printed order.
// CR 612.2, text_words.go, ADR 0160's 2026-10-15 amendment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe371e3a-976f-41d8-ad58-4b3b61415d71",
		Name:         "Crystal Spray",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetSpellOrPermanent("target spell or permanent", nil, nil),
		OnResolve: changeTextOfTarget(textEitherWords, true, func(item *game.StackItem) func(*game.Game) error {
			controller := item.Controller
			return func(g *game.Game) error { return g.DrawNForEffect(controller, 1) }
		}),
	})
}

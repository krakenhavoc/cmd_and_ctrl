package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Tilling — Instant {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"Mill four cards, then you may return a permanent card from among them
//	 to your hand."
//
// Cache Grab's mill and pick (rfSpellBMillThenMayTakePermanent) with
// nothing after the pick. The pick is over the cards that reached the
// graveyard (CR 400.7), up to one.
func init() {
	Register(Spec{
		OracleID:     "79bc2c6d-b03d-4ceb-b468-293378ec73ae",
		Name:         "Midnight Tilling",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return rfSpellBMillThenMayTakePermanent(ctx,
				"Midnight Tilling — return a permanent card milled this way to your hand",
				midnightTillingMilled)
		},
	})
}

// midnightTillingMilled is the mill's keyed "then". An on-disk key: never
// renamed, never reused.
var midnightTillingMilled = rfSpellBTakeFromMilledTail("mill-then-take/midnight-tilling", func(*Context) error { return nil })

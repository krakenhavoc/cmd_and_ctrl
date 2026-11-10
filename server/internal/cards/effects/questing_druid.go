package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Questing Druid // Seek the Beast — Creature — Human Druid {1}{G}, 1/1 //
// Instant — Adventure {1}{R} (Wilds of Eldraine, #4471):
//
//	Questing Druid — "Whenever you cast a spell that's white, blue,
//	 black, or red, put a +1/+1 counter on this creature."
//	Seek the Beast — "Exile the top two cards of your library. Until your
//	 next end step, you may play those cards."
//
// The trigger reads the spell's colours as it is cast (a multicolour spell
// that is green and red counts once: it is one trigger per spell). The
// Adventure is Opera Love Song's exile with a window that ends as your
// next end step begins (game.UntilYourNextEndStep).
//
// No simplification.
const questingDruidOracleID = "555df32f-f8e9-49af-bc4a-1194ea325a97"

func init() {
	Register(Spec{
		OracleID:     questingDruidOracleID,
		Name:         "Questing Druid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(Or(OfColor("W"), OfColor("U"), OfColor("B"), OfColor("R"))),
				"Questing Druid — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
	Register(Spec{
		OracleID:     questingDruidOracleID + "#1",
		Name:         "Seek the Beast",
		Completeness: CompletenessFull,
		OnResolve:    func(_ *game.StackItem, ctx *Context) error { return ExileTopNUntilYourNextEndStep(ctx, 2) },
	})
}

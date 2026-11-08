package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keen Buccaneer — Creature — Octopus Pirate {2}{U}, 2/3 (Aetherdrift,
// #4464):
//
//	"Vigilance
//	 Exhaust — {1}{U}: Draw a card, then discard a card. Put a +1/+1
//	 counter on this creature. (Activate each exhaust ability only once.)"
//
// Camera Launcher's exhaust (CR 702.177a: "Activate only once"), keyed to
// the object, with Merfolk Looter's loot first. The counter waits for the
// discard the loot asks (CR 608.2c), and lands only while the Buccaneer
// is still the permanent that activated: one that left and came back in
// response is a new object (CR 400.7). Vigilance is the engine's
// (CR 702.20).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8cadf9c0-9a63-4cde-a6be-0ae86890c831",
		Name:            "Keen Buccaneer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {1}{U}: Draw a card, then discard a card. Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Draws: 1, Discards: 1, SourceCounters: 1},
			Exhaust: true,
			Cost:    ManaCost("{1}{U}"),
			Effect: Do(
				Instruction(func(ctx *Context) error { return lootOne(ctx.Game, ctx.Item, 1) }),
				Instruction(plusOneCounterOnThisIfStillHere),
			),
		}},
	})
}

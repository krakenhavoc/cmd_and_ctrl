package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pacesetter Paragon — Creature — Human Pilot {2}{R}, 2/3 (Aetherdrift,
// #4464):
//
//	"Exhaust — {2}{R}: Put a +1/+1 counter on this creature. It gains
//	 double strike until end of turn. (Activate each exhaust ability only
//	 once.)"
//
// Prowcatcher Specialist's exhaust (CR 702.177a: "Activate only once"),
// keyed to the object, so a Paragon back from a flicker may use it again
// (CR 400.7). The counter stays; double strike ends with the turn. Both
// land on the Paragon only while it is still the permanent that
// activated: one that left and came back in response gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1acb289f-fb7f-4d83-aaef-4e7b9462c593",
		Name:         "Pacesetter Paragon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Exhaust — {2}{R}: Put a +1/+1 counter on this creature. It gains double strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 1},
			Exhaust: true,
			Cost:    effects.ManaCost("{2}{R}"),
			// The keyword waits for any prompt the counter asks (CR 608.2c).
			Effect: effects.Do(
				effects.Instruction(effects.PlusOneCounterOnThisIfStillHere),
				effects.Instruction(func(ctx *effects.Context) error {
					if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: ctx.Item.SourceCardID, Keywords: []string{"double strike"},
						Label: "Pacesetter Paragon — double strike"}.Apply(ctx)
				}),
			),
		}},
	})
}

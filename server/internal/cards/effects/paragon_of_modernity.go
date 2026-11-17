package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paragon of Modernity — Artifact Creature — Angel Warrior {4}, 2/2
// (Streets of New Capenna):
//
//	"Flying
//	 {3}: This creature gets +1/+1 until end of turn. If exactly three
//	 colors of mana were spent to activate this ability, put a +1/+1
//	 counter on it instead."
//
// The colours of the mana spent on the activation are read as it
// resolves (StackItem.Paid, #3855): exactly three distinct colours
// (colourless is no colour) put the counter. Paragon declares
// ReadsOwnActivations, so a payment waived outside the pool asks what was
// paid. A payment the engine did not see counts as fewer colours:
// weaker, never stronger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1c62704-1f20-48cf-8588-47e7fca85f7b",
		Name:            "Paragon of Modernity",
		ReadsManaSpent:  game.ReadsOwnActivations,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{3}: This creature gets +1/+1 until end of turn. If exactly three colors of mana were spent to activate this ability, put a +1/+1 counter on it instead.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}"),
			Effect: s82SncB11PumpOrCounter("Paragon of Modernity — +1/+1 until end of turn",
				func(ctx *Context) bool { return ctx.ColorsSpentCount() == 3 }),
		}},
	})
}

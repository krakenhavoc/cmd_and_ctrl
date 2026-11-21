package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "c1c62704-1f20-48cf-8588-47e7fca85f7b",
		Name:            "Paragon of Modernity",
		ReadsManaSpent:  game.ReadsOwnActivations,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This creature gets +1/+1 until end of turn. If exactly three colors of mana were spent to activate this ability, put a +1/+1 counter on it instead.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}"),
			Effect: effects.S82SncB11PumpOrCounter("Paragon of Modernity — +1/+1 until end of turn",
				func(ctx *effects.Context) bool { return ctx.ColorsSpentCount() == 3 }),
		}},
	})
}

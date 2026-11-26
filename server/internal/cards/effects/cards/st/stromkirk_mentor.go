package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stromkirk Mentor — Creature — Vampire Soldier {3}{B}, 4/2 (Shadows over
// Innistrad, #5543):
//
//	"When this creature enters, put a +1/+1 counter on another target
//	 Vampire you control."
//
// "Another" is object identity (it can't target itself), and the word
// Vampire is a word of the card's text (CR 612.2). The counter goes on the
// target if it is still legal as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17cbeec4-b180-48e9-9156-42271f918a9d",
		Name:         "Stromkirk Mentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Stromkirk Mentor — put a +1/+1 counter on another target Vampire you control",
				effects.PlusOneCounterOnTheTarget),
				effects.Another(effects.TargetCreature("another target Vampire you control", effects.YouControl(), effects.Subtype("Vampire")))),
		},
	})
}

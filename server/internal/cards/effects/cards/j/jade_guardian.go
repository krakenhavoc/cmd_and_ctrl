package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jade Guardian — Creature — Merfolk Shaman {3}{G}, 2/2 (Ixalan, #5532):
//
//	"Hexproof
//	 When this creature enters, put a +1/+1 counter on target Merfolk you
//	 control."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and is
// any Merfolk you control, the Guardian itself included; it is judged
// again as the trigger resolves (CR 608.2b), so a Merfolk that left or
// lost the type gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9128b952-bd36-4a3a-9566-34d999305991",
		Name:            "Jade Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Jade Guardian — put a +1/+1 counter on target Merfolk you control", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target Merfolk you control", effects.Subtype("Merfolk"), effects.YouControl())),
		},
	})
}

package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keen-Eyed Raven — Creature — Bird {2}{W}, 2/2 (Assassin's Creed,
// #5466):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on another target
//	 creature you control."
//
// "Another" is object identity (Another excludes the source), so the
// Raven never offers itself; with no other creature the trigger is
// removed for want of a target (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e42450c7-a9de-412f-b439-76f2f9b6031b",
		Name:            "Keen-Eyed Raven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Keen-Eyed Raven — put a +1/+1 counter on another target creature you control",
					effects.PlusOneCounterOnChosenTarget),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}

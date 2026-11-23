package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sterling Supplier — Creature — Bird Soldier {4}{W}, 3/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on another target
//	 creature you control."
//
// Dr. McCoy, Surly Surgeon's enters trigger. "Another" excludes the
// Supplier itself; with no other creature to target the trigger is
// removed (CR 603.3d).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Sterling Supplier — put a +1/+1 counter on another target creature you control", effects.PlusOneCounterOnTheTarget)
	enters.Targets = effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:        "ab23cca2-b59e-4d34-80de-98bd08985275",
		Name:            "Sterling Supplier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Beaver — Creature — Beaver Mount {3}{G}, 4/4:
//
//	"Vigilance
//	 Whenever this creature attacks while saddled, put a +1/+1 counter on
//	 target creature that saddled it this turn.
//	 Saddle 3"
//
// The target clause is SaddledThis (ADR 0186 section 3): the creatures
// tapped to pay THIS Beaver's saddle cost this turn, read from the turn's
// crew records by the Beaver's object and still on the battlefield as the
// same objects. So a creature that saddled another Mount is not offered, a
// saddler that left is not, and the trigger is removed from the stack for
// want of a target when none is left (CR 603.3d). The counter goes on the
// announced creature if it is still a legal target at resolution (CR
// 608.2b); a saddler flickered in response is a new object and the
// trigger fizzles.
//
// No simplification.
func init() {
	attack := effects.AttacksWhileSaddled("Giant Beaver — put a +1/+1 counter on target creature that saddled it", effects.B36CounterOnChosenAnimal)
	attack.Targets = effects.SaddledThis(effects.TargetCreature("target creature that saddled it this turn"))

	effects.Register(effects.Spec{
		OracleID:        "2a4b42b3-6a95-41e1-9741-b35723d8da73",
		Name:            "Giant Beaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Saddle(3)},
		Triggered:       []game.TriggeredAbility{attack},
	})
}

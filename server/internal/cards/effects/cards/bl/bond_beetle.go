package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bond Beetle — Creature — Insect {G}, 0/1 (Magic 2013, #5568):
//
//	"When this creature enters, put a +1/+1 counter on target creature."
//
// Cloudbound Moogle's enters trigger: any creature, the Beetle itself
// included, and nothing if the target is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9353f6dd-9a08-4e55-9b07-3e68d7d97ef4",
		Name:         "Bond Beetle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Bond Beetle — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
	})
}

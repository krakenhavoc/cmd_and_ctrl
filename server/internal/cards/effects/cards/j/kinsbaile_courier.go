package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kinsbaile Courier — {2}{W} Creature — Kithkin Soldier 2/1, Commander Legends (#5507):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 Encore {2}{W}"
//
// Bond Beetle's enters trigger (any creature, itself included, and nothing if
// the target is gone, CR 608.2b) and the shared Encore constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92c0346d-c375-4851-9806-5d3b9b83425f",
		Name:         "Kinsbaile Courier",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Encore("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Kinsbaile Courier — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
	})
}

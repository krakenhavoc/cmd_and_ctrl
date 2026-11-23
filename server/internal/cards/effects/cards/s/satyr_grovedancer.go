package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Satyr Grovedancer — Creature — Satyr Shaman {1}{G}, 1/1 (Journey into Nyx, #5558):
//
//	"When this creature enters, put a +1/+1 counter on target creature."
//
// Ironpaw Aspirant's enters counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49f3bd5a-9c2a-45b8-aa10-9c6471028a0f",
		Name:         "Satyr Grovedancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Satyr Grovedancer — put a +1/+1 counter on target creature", effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature")),
		},
	})
}

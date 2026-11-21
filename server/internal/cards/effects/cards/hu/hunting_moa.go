package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunting Moa — Creature — Bird Beast, {2}{G}, 3/2:
//
//	"Echo {2}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or dies, put a +1/+1 counter on target creature."
//
// "Enters or dies" is one ability with two trigger conditions.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "102d73bc-94ed-4a57-9872-6f7b9bf282f9",
		Name:         "Hunting Moa",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Hunting Moa", "{2}{G}"),
			effects.Targeting(effects.WhenThisEntersOrDies("Hunting Moa — put a +1/+1 counter on target creature", effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("target creature")),
		},
	})
}

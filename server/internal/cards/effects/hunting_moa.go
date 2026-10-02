package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunting Moa — Creature — Bird Beast, {2}{G}, 3/2:
//
//	"Echo {2}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or dies, put a +1/+1 counter on target creature."
//
// "Enters or dies" is one ability with two trigger conditions.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "102d73bc-94ed-4a57-9872-6f7b9bf282f9",
		Name:         "Hunting Moa",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Hunting Moa", "{2}{G}"),
			Targeting(WhenThisEntersOrDies("Hunting Moa — put a +1/+1 counter on target creature", putPlusOneCounterOnEachLegalTarget),
				TargetCreature("target creature")),
		},
	})
}

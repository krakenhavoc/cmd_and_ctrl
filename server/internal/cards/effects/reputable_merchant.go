package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reputable Merchant — Creature — Human Citizen {2/W}{2/B}{2/G}, 2/2
// (Tarkir: Dragonstorm, #4463):
//
//	"When this creature enters or dies, put a +1/+1 counter on target
//	 creature you control."
//
// Hunting Moa's one ability with two trigger conditions, narrowed to a
// creature you control. On entry the Merchant may target itself; when it
// dies it is gone, so another creature takes the counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62f09a5b-d7fb-44f4-9fab-719185b735ab",
		Name:         "Reputable Merchant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrDies("Reputable Merchant — put a +1/+1 counter on target creature you control",
				putPlusOneCounterOnEachLegalTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

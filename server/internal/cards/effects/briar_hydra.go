package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Briar Hydra — Creature — Plant Hydra {5}{G}, 6/6 (Dominaria United):
//
//	"Trample
//	 Domain — Whenever this creature deals combat damage to a player,
//	 put X +1/+1 counters on target creature you control, where X is the
//	 number of basic land types among lands you control."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and
// may be the Hydra itself; X is the shared domain count (DomainCount,
// #5666) read as the ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Briar Hydra — put X +1/+1 counters on target creature you control, where X is the number of basic land types among lands you control"
	Register(Spec{
		OracleID:        "caaedd67-e074-47c4-a956-da88a01f8996",
		Name:            "Briar Hydra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsCombatDamageToAPlayer(label, dmuB11DomainCounters()),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

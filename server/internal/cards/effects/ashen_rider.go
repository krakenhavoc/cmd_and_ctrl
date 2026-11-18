package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashen Rider — Creature — Archon {4}{W}{W}{B}{B}, 5/5:
//
//	"Flying
//	 When this creature enters or dies, exile target permanent."
//
// One printed trigger, so one registered row that watches both events
// (WhenThisEntersOrDies). The target is chosen as the trigger goes on the
// stack and is exiled as it resolves; a target that has left is illegal
// and the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "455f66eb-3e33-4b51-9ec1-32288bdd3a68",
		Name:            "Ashen Rider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrDies("Ashen Rider — exile target permanent", ExileFirstTarget), TargetPermanent("target permanent")),
		},
	})
}

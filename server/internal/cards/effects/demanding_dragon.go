package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demanding Dragon — Creature — Dragon {3}{R}{R}, 5/5:
//
//	"Flying
//	 When this creature enters, it deals 5 damage to target opponent
//	 unless that player sacrifices a creature of their choice."
//
// A targeted punisher: the targeted opponent chooses between the damage
// and sacrificing a creature (the Dragon is the damage's source). With no
// creature to sacrifice the damage is dealt without asking. The body is in
// s82_rpr_b33_helpers.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8f1e3bd-bbf0-4750-b054-03553fc61850",
		Name:            "Demanding Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Demanding Dragon — 5 damage to target opponent unless they sacrifice a creature", demandingDragonAsk),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

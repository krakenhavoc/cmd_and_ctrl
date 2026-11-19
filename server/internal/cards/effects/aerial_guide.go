package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerial Guide — Creature — Drake {2}{U}, 2/2 (Foundations Jumpstart):
//
//	"Flying
//	 Whenever this creature attacks, another target attacking creature gains
//	 flying until end of turn."
//
// Helitrooper's attack trigger.
//
// No simplification.
func init() {
	const label = "Aerial Guide — another attacking creature gains flying"
	attack := WheneverThisAttacks(label, firstTargetGainsUntilEndOfTurn(0, label, "flying"))
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	Register(Spec{
		OracleID:        "b56c0846-e465-4251-93ea-2e3e7be8dffc",
		Name:            "Aerial Guide",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}

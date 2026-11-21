package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	attack := effects.WheneverThisAttacks(label, effects.FirstTargetGainsUntilEndOfTurn(0, label, "flying"))
	attack.Targets = effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))
	effects.Register(effects.Spec{
		OracleID:        "b56c0846-e465-4251-93ea-2e3e7be8dffc",
		Name:            "Aerial Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}

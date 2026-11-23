package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seal Away — Enchantment {1}{W} (Dominaria, #5530):
//
//	"Flash
//	 When this enchantment enters, exile target tapped creature an opponent controls until this enchantment leaves the battlefield."
//
// Super Villain Lockup's shape: flash, and the tapped creature is exiled
// until this enchantment leaves the battlefield (CR 610.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8909015-ea49-47cb-8d37-653904f965bb",
		Name:            "Seal Away",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets: effects.TargetCreature("target tapped creature an opponent controls",
					effects.OpponentControls(), effects.B751Tapped()),
				Key:    "Seal Away — exile target tapped creature an opponent controls",
				Effect: effects.ExileChosenTargetUntilThisLeaves("Seal Away — the exiled creature returns when the enchantment leaves the battlefield"),
			},
		},
	})
}

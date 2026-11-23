package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Super Villain Lockup — {1}{W} Enchantment (Marvel Super Heroes, #4455):
//
//	"Flash
//	 When this enchantment enters, exile target tapped creature an
//	 opponent controls until this enchantment leaves the battlefield."
//
// Oblivion Ring's one-shot (CR 610.3, exileChosenTargetUntilThisLeaves):
// the return happens the moment the enchantment leaves, and if it
// leaves before the trigger resolves nothing is exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "192e150a-96b0-42d6-9d39-03d0bed61f41",
		Name:            "Super Villain Lockup",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets: effects.TargetCreature("target tapped creature an opponent controls",
					effects.OpponentControls(), effects.B751Tapped()),
				Key:    superVillainLockupLabel,
				Effect: effects.ExileChosenTargetUntilThisLeaves("Super Villain Lockup — the exiled creature returns when the enchantment leaves the battlefield"),
			},
		},
	})
}

const superVillainLockupLabel = "Super Villain Lockup — exile target tapped creature an opponent controls"

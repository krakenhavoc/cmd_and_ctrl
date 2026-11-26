package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fairgrounds Warden — Creature — Dwarf Soldier {2}{W}, 1/3 (Kaladesh, #5539):
//
//	"When this creature enters, exile target creature an opponent controls
//	 until this creature leaves the battlefield."
//
// Banisher Priest's Oblivion Ring tail over a creature an opponent controls:
// the card returns under its owner's control when the Warden leaves (CR
// 610.3c); if the Warden has already left when the trigger resolves nothing
// is exiled (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "388b7b0f-b26d-4de9-bfa6-8c3cbcc2e284",
		Name:         "Fairgrounds Warden",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Fairgrounds Warden — exile target creature an opponent controls until this creature leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Fairgrounds Warden — the exiled creature returns when Fairgrounds Warden leaves the battlefield")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

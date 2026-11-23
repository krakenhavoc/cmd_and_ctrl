package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bounding Krasis — Creature — Fish Lizard {1}{G}{U}, 3/3 (Magic Origins, #5549):
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 When this creature enters, you may tap or untap target creature."
//
// Flash, and Trickster Mage's tap-or-untap choice on an entry trigger:
// the creature is targeted as the trigger goes on the stack and the
// controller picks tap, untap or neither as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c6736bac-314b-48af-8722-c16a0d66affa",
		Name:            "Bounding Krasis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Bounding Krasis — you may tap or untap target creature", effects.NemMayTapOrUntapTheTarget("Bounding Krasis")),
				effects.TargetCreature("target creature")),
		},
	})
}

package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fervent Cathar — {2}{R} Creature — Human Knight 2/1 (Avacyn Restored, #5570):
//
//	"Haste
//	 When this creature enters, target creature can't block this turn."
//
// Voldaren Duelist's haste and enters trigger: the target can't
// block this turn.
//
// No simplification.
func init() {
	const label = "Fervent Cathar — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:        "94bcae2f-86a6-4d38-b8e1-5ebf69dc7cbc",
		Name:            "Fervent Cathar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
					effects.TargetCreature("target creature")),
				effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
		},
	})
}

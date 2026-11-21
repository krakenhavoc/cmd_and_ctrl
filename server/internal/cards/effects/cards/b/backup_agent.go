package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Backup Agent (Streets of New Capenna, #5491):
//
//	When this creature enters, put a +1/+1 counter on target creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df9a3fe0-aca8-4adb-9ef6-45616025925b",
		Name:         "Backup Agent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Backup Agent — put a +1/+1 counter on target creature",
					effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
		},
	})
}

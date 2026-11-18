package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Backup Agent (Streets of New Capenna, #5491):
//
//	When this creature enters, put a +1/+1 counter on target creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df9a3fe0-aca8-4adb-9ef6-45616025925b",
		Name:         "Backup Agent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Backup Agent — put a +1/+1 counter on target creature",
					PlusOneCounterOnTheTarget),
				TargetCreature("target creature")),
		},
	})
}

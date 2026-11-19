package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kazandu Nectarpot — Creature — Insect {1}{G}, 1/3:
//
//	"Landfall — Whenever a land you control enters, you gain 1 life."
//
// Eumidian Terrabotanist's landfall row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "263b526c-8b81-44ee-a7c4-5c24bf9b42a4",
		Name:         "Kazandu Nectarpot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Kazandu Nectarpot — you gain 1 life (landfall)", game.LandfallPayoff{LifeGain: 1}, Do(GainLife{Amount: 1})),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eumidian Terrabotanist — Creature — Insect Druid {1}{G}, 2/3 (Edge of
// Eternities, #4461):
//
//	"Landfall — Whenever a land you control enters, you gain 1 life."
//
// Courser of Kruphix's landfall row on its own: one trigger for each land
// that enters under your control, an opponent's land none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92226fd2-ad93-4722-89a0-ca88ea03e1b4",
		Name:         "Eumidian Terrabotanist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Eumidian Terrabotanist — you gain 1 life (landfall)", game.LandfallPayoff{LifeGain: 1}, Do(GainLife{Amount: 1})),
		},
	})
}

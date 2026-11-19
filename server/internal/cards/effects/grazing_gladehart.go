package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grazing Gladehart — Creature — Antelope {2}{G}, 2/2 (Foundations
// Jumpstart):
//
//	"Landfall — Whenever a land you control enters, you may gain 2
//	 life."
//
// Landfall with a CR 603.5 "you may"; the payoff is declared for the
// bot. No simplification.
func init() {
	Register(Spec{
		OracleID:     "f19f28e5-9cad-4398-b2d4-9e7fefb23cb4",
		Name:         "Grazing Gladehart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(LandfallPaying("Grazing Gladehart — you may gain 2 life (landfall)",
				game.LandfallPayoff{LifeGain: 2}, Do(GainLife{Amount: 2})),
				"Grazing Gladehart — gain 2 life?"),
		},
	})
}

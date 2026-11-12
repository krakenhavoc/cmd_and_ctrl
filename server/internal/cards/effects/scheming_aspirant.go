package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scheming Aspirant — Creature — Phyrexian Advisor {1}{B}, 1/3
// (Phyrexia: All Will Be One, #5653):
//
//	"Whenever you proliferate, each opponent loses 2 life and you gain 2
//	 life."
//
// WheneverYouProliferate: one trigger per proliferate its controller
// finishes, whatever was chosen (CR 701.34a), so "proliferate twice"
// drains twice. Life loss, not damage, and the gain is 2 however many
// opponents lost life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf4a3b1a-4b87-473e-be1f-3aa3da45c28e",
		Name:         "Scheming Aspirant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouProliferate("Scheming Aspirant — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}

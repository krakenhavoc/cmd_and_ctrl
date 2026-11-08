package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Candy Trail — Artifact — Food Clue {1} (Wilds of Eldraine, #4471):
//
//	"When this artifact enters, scry 2.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life and draw a card."
//
// The entry scry is Giant's Boulder's trigger. The card prints one merged
// ability in place of the usual Food and Clue sacrifices: the life and the
// card come together, Bagel and Schmear's Nosh (gainThreeLifeAndDrawAbility).
// It answers nothing on its own, so it declares AnswerValue (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c8e4f29-53e0-4242-8ef8-7a9476dce178",
		Name:         "Candy Trail",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Candy Trail — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{gainThreeLifeAndDrawAbility("{2}, {T}, Sacrifice this artifact: You gain 3 life and draw a card.")},
	})
}

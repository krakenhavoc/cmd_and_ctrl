package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wakandan Tusker — Creature — Elephant {1}{G/W}, 2/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Whenever this creature becomes tapped, you gain 1 life and scry 1."
//
// Dionus's "becomes tapped" (ThisBecameTapped, CR 701.26a): any tap, for
// mana, by an effect or by attacking, and not a vigilance attacker. The
// life is gained first and the scry is asked after it, in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e8ffb63-1a7b-4ebe-839d-d5192522db65",
		Name:         "Wakandan Tusker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Wakandan Tusker — you gain 1 life and scry 1", Do(GainLife{Amount: 1}, Scry{N: 1})),
		},
	})
}

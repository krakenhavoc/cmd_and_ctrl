package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wanderbrine Preacher — Creature — Merfolk Cleric {1}{W}, 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever this creature becomes tapped, you gain 2 life."
//
// Compassionate Healer's trigger with two life and no scry: "becomes
// tapped" is the tap event and the attack that taps it, so a vigilance
// attack is not one and crewing or paying a cost is (CR 701.26a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b4dcdb1-0017-494f-8d83-6706b188e575",
		Name:         "Wanderbrine Preacher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Wanderbrine Preacher — you gain 2 life",
				Do(GainLife{Amount: 2})),
		},
	})
}

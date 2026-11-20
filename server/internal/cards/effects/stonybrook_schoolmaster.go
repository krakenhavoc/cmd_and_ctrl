package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stonybrook Schoolmaster — Creature — Merfolk Wizard {2}{W}, 1/2
// (Morningtide, #5592):
//
//	"Whenever this creature becomes tapped, you may create a 1/1 blue
//	 Merfolk Wizard creature token."
//
// Attentive Sunscribe's "becomes tapped" trigger (ThisBecameTapped, which
// watches a tap and an attack, CR 701.26) with an optional token
// (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ef67bae-4af3-47b9-95d2-08f3c5a0b39d",
		Name:         "Stonybrook Schoolmaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Stonybrook Schoolmaster — create a 1/1 blue Merfolk Wizard token",
				Do(CreateToken{Template: TokenCard("1/1 blue Merfolk Wizard"), N: 1})),
				"Stonybrook Schoolmaster — create a 1/1 blue Merfolk Wizard token?"),
		},
	})
}

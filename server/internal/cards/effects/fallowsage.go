package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallowsage — Creature — Merfolk Wizard {3}{U}, 2/2:
//
//	"Whenever this creature becomes tapped, you may draw a card."
//
// "Becomes tapped" is the tap event and the attack that taps it
// (ThisBecameTapped), as Attentive Sunscribe reads it. The draw is
// optional, asked as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9d018ff-de27-4bf6-9cc6-9d9a46db3eb9",
		Name:         "Fallowsage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OptionalDraw(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Fallowsage — you may draw a card",
				Do(DrawCards{N: 1})), "Draw a card?", 1),
		},
	})
}

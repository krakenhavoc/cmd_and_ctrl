package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silvergill Peddler — Creature — Merfolk Citizen {2}{U}, 2/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever this creature becomes tapped, draw a card, then discard a
//	 card."
//
// Attentive Sunscribe's "becomes tapped" (the tap event and an attack that
// taps it, #4534) with the shared loot body: the draw lands first and the
// discard is the controller's choice among the whole hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08688903-d9cc-4d57-ba91-14a8523551b9",
		Name:         "Silvergill Peddler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Silvergill Peddler — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1) }),
		},
	})
}

package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b9d018ff-de27-4bf6-9cc6-9d9a46db3eb9",
		Name:         "Fallowsage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Fallowsage — you may draw a card",
				effects.Do(effects.DrawCards{N: 1})), "Draw a card?", 1),
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Underworld Dreams — Enchantment {B}{B}{B} (EDHREC rank 1436):
//
//	"Whenever an opponent draws a card, this enchantment deals 1
//	 damage to that player."
//
// The wheel deck's payoff. Scrawling Crawler's second ability with
// damage in place of life loss: one trigger per card drawn, for any
// opponent's draw from any source, and the drawing player rides the
// trigger's closure as a copied ID. Damage, so it is the
// enchantment's — a damage doubler sees it and a prevention shield
// stops it, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "967cf377-ae26-464d-85ac-8448b5a911f7",
		Name:         "Underworld Dreams",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				drawer := ev.Actor
				return game.NewTriggeredItem(source, "Underworld Dreams — 1 damage to that player",
					func(g *game.Game, item *game.StackItem) error {
						if g.PlayerByIDForEffect(drawer) == nil {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: drawer, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}

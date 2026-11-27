package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karma — Enchantment {2}{W}{W} (Eighth Edition, S82 OLD batch 611, #5607):
//
//	"At the beginning of each player's upkeep, this enchantment deals damage to
//	 that player equal to the number of Swamps they control."
//
// Roiling Vortex's upkeep trigger: it hits whoever's upkeep it is, Karma's
// controller included. The Swamps counted are the lands that player controls
// with the Swamp land type as the trigger resolves (CR 205.3i, read through the
// layers), so a land that has become a Swamp counts; no Swamps, no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fac4da47-f0b2-4b57-9703-e0ed100d3499",
		Name:         "Karma",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Karma — damage to that player equal to the number of Swamps they control",
				func(g *game.Game, item *game.StackItem) error {
					p := effects.TriggeringActor(item)
					if p == uuid.Nil {
						return nil
					}
					swamp := "Swamp"
					if karma, ok := g.LookupCardForEffect(item.SourceCardID); ok {
						swamp = karma.TextSubtype("Swamp") // a word of Karma's text (CR 612.2)
					}
					swamps := effects.CountControlled(g, p, func(c game.Card) bool { return c.IsLand() && c.HasSubtype(swamp) })
					return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: swamps}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

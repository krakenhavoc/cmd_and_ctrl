package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flesh Reaver — Creature — Phyrexian Horror {1}{B}, 4/4 (Dominaria
// Remastered):
//
//	"Whenever this creature deals damage to a creature or opponent, this
//	 creature deals that much damage to you."
//
// One trigger per damage event the Reaver is the source of, combat or not,
// whose recipient is a creature on the battlefield (ThisDealtDamageToACreature)
// or one of its controller's opponents (ThisDealtDamageToAnOpponent). The
// amount is the damage actually dealt, carried on the trigger as it
// triggers (prevented damage does not trigger), and the Reaver deals it to
// its controller as it last existed if it has left (CR 608.2h). Damage to a
// planeswalker or to you does not trigger it.
//
// No simplification.
func init() {
	const label = "Flesh Reaver — it deals that much damage to you"
	effects.Register(effects.Spec{
		OracleID:     "a143fdfe-f1aa-4eab-8c81-8c72a4dc1695",
		Name:         "Flesh Reaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.ThisDealtDamageToACreature(ev, source, ch, g) || effects.ThisDealtDamageToAnOpponent(ev, source, ch, g)
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.DamageController{N: item.Params.Amount}.Apply(ctx)
			},
		}},
	})
}

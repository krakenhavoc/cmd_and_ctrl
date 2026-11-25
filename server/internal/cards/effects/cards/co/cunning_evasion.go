package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Evasion — Enchantment {1}{U} (Modern Horizons):
//
//	"Whenever a creature you control becomes blocked, you may return it to its
//	 owner's hand."
//
// The trigger watches the becomes-blocked event (CR 509.1h, 509.5): the blocked
// attacker is the event's target and must be a creature its controller controls
// as it becomes blocked. "You may" is asked as the ability resolves, and the
// creature returns only while it is still the same object on the battlefield
// (CR 400.7, 608.2b); one that has left already returns nothing.
//
// No simplification.
const cunningEvasionLabel = "Cunning Evasion — you may return it to its owner's hand"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e56606d-562e-4c77-9769-213997fea28e",
		Name:         "Cunning Evasion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && ev.Kind == game.EventBecomesBlocked && c.IsCreature() && c.Controller == source.Controller
			}, cunningEvasionLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				obj := ctx.Trigger().Object
				if obj == nil {
					return nil
				}
				if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
					return nil
				}
				return effects.BounceToHand{Target: obj.Ref().ID}.Apply(ctx)
			}), cunningEvasionLabel+"?"),
		},
	})
}

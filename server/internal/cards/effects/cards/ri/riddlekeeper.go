package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riddlekeeper — Creature — Homunculus {2}{U}, 1/4 (Commander 2011, #5574):
//
//	"Whenever a creature attacks you or a planeswalker you control, that
//	 creature's controller mills two cards."
//
// One trigger per attacking creature (CR 508.1m: a creature's attack is its own
// event), read from the attack event: the attacked player is Riddlekeeper's
// controller, or the attacked permanent is a planeswalker they control. The
// mill is the attacker's controller's, read from the attacking creature as it
// is when the trigger resolves or as it last stood (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5e37c6e-ae42-4c98-ba20-840ea3c4a6ab",
		Name:         "Riddlekeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, riddlekeeperAttacked,
				"Riddlekeeper — that creature's controller mills two cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					attacker := item.Trigger.Event.CardID
					controller := item.Trigger.Event.Actor
					if c, ok := g.LookupCardForEffect(attacker); ok {
						controller = c.Controller
					}
					if controller == uuid.Nil {
						return nil
					}
					return effects.MillCards{Player: controller, N: 2}.Apply(ctx)
				}),
		},
	})
}

// riddlekeeperAttacked is "a creature attacks you or a planeswalker you
// control" (CR 508.3a), AttacksYouOrYourPlaneswalker over the attack's target.
func riddlekeeperAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil && effects.AttacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
}

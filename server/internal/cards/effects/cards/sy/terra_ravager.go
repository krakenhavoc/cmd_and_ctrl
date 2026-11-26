package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Terra Ravager — Creature — Elemental Beast {2}{R}{R}, 0/4 (Commander 2013,
// S82 older sets):
//
//	"Whenever this creature attacks, it gets +X/+0 until end of turn, where X
//	 is the number of lands defending player controls."
//
// X is read as the trigger resolves, for the defending player the attack named
// (the player attacked, or the controller of the planeswalker or battle
// attacked), so a land played or destroyed in response counts. The attack's
// defender is fixed by the event, so a creature that was re-pointed since still
// reads the player it attacked when it triggered (CR 508.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7686204-0433-48cf-bbfb-5d32b6a25cc3",
		Name:         "Terra Ravager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			}, "Terra Ravager — it gets +X/+0 until end of turn, where X is the number of lands defending player controls",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if defender == uuid.Nil {
						return nil
					}
					x := effects.CountControlled(g, defender, func(c game.Card) bool { return c.IsLand() })
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Terra Ravager — +X/+0"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

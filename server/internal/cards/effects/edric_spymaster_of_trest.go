package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edric, Spymaster of Trest — 2/2 Legendary Creature — Elf Rogue
// for {1}{G}{U}:
//
//	"Whenever a creature deals combat damage to one of your
//	opponents, its controller may draw a card."
//
// S19 sub-PR 7: the political one. ANY creature qualifies (not just
// Edric's controller's), the damaged player must be one of Edric's
// controller's opponents, and the "you may" belongs to the
// creature's controller — so both the prompt (OptionalPrompt.Chooser
// override) and the draw go to that player, not to Edric's owner.
// The trigger is still Edric's ability, so it sits on the stack
// under Edric's controller (APNAP-ordered with their other triggers)
// and its resolution draws for the creature's controller, which
// the combat-damage event carries in Actor.
func init() {
	Register(Spec{
		OracleID: "9a1de7e4-9930-4db3-a8f3-d146d0abf38b",
		Name:     "Edric, Spymaster of Trest",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
					return false
				}
				if ev.Target == source.Controller {
					return false // damage to Edric's own controller is not "one of your opponents"
				}
				if p := g.PlayerByIDForEffect(ev.Target); p == nil {
					return false
				}
				src, ok := g.LookupCardForEffect(ev.Source)
				return ok && src.IsCreature()
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				// Actor is the dealing creature's controller, stamped
				// at emit time — still valid if the creature has since
				// died to simultaneous combat damage.
				drawer := ev.Actor
				if drawer == uuid.Nil {
					return nil
				}
				return game.NewTriggeredItem(source, "Edric — attacking creature's controller draws a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: drawer, N: 1}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Edric, Spymaster of Trest — draw a card?",
				Chooser: func(ev game.Event, _ *game.Card, _ *game.Game) uuid.UUID {
					return ev.Actor
				},
			},
		}},
	})
}

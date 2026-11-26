package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ash Zealot — Creature — Human Warrior {R}{R}, 2/2 (Return to Ravnica, #5566):
//
//	"First strike, haste
//	 Whenever a player casts a spell from a graveyard, this creature deals 3
//	 damage to that player."
//
// River Kelpie's "a player casts a spell from a graveyard" (EventCast whose
// OldZone is the graveyard), for any player. The damage goes to the caster,
// the event's actor, and this creature is the source (CR 120.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab9f03b4-eb67-4709-904d-abf4d491f0ee",
		Name:            "Ash Zealot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCast && ev.OldZone == game.ZoneGraveyard
			}, "Ash Zealot — 3 damage to the player who cast a spell from a graveyard",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || g.PlayerByIDForEffect(item.Trigger.Event.Actor) == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 3}.Apply(ctx)
				}),
		},
	})
}

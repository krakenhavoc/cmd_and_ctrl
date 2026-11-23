package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elminster — Legendary Planeswalker — Elminster {3}{W}{U}, loyalty 5
// (Commander Legends: Battle for Baldur's Gate, S82 GAP #5948):
//
//	"Whenever you scry, the next instant or sorcery spell you cast this turn
//	 costs {X} less to cast, where X is the number of cards looked at while
//	 scrying this way.
//	 +2: Draw a card, then scry 2.
//	 −3: Exile the top card of your library. Create a number of 1/1 blue Faerie
//	     Dragon creature tokens with flying equal to that card's mana value.
//	 Elminster can be your commander."
//
// The scry trigger fires after a completed scry (EventScry, nothing is looked at
// from an empty library) and stores the size of that scry, the event's LookedAt
// (not the cards that moved), in Kaza's one-use promise for the next instant or
// sorcery this turn (#1852); each scry is its own promise. The +2 draws, then
// scries after the draw, so the draw is first and the scry is of the library as
// it stands (and itself triggers Elminster, so a +2 makes the next instant or
// sorcery cost {2} less). The -3 exiles the top card (an exile, not a mill) and
// makes tokens equal to its mana value as it was in exile (a land is 0, X is 0);
// with an empty library nothing is made. "Can be your commander" is deck
// construction, read from the oracle text by the deck checker.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "56773795-bd8b-4ac4-8f2c-a9d6275989bb",
		Name:            "Elminster",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventScry, effects.ByYou,
				"Elminster — the next instant or sorcery spell you cast this turn costs {X} less to cast",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					x := item.Trigger.Event.LookedAt
					if x <= 0 {
						return nil
					}
					return effects.GrantNextSpellPromise{From: "Elminster", Promise: game.NextSpellPromise{
						Filter: game.PermissionFilter{InstantOrSorceryOnly: true},
						Reduce: x,
						Text:   "The next instant or sorcery spell you cast this turn costs {X} less to cast.",
					}}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Draw a card, then scry 2.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Draws: 1, Scry: 2},
				Effect:  effects.Do(effects.DrawCards{N: 1}, effects.Scry{N: 2}),
			},
			{
				Label: "−3: Exile the top card of your library. Create a number of 1/1 blue Faerie Dragon creature tokens with flying equal to that card's mana value.",
				Cost:  effects.LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{Player: item.Controller, N: 1, To: game.ZoneExile, ThenTail: ctx.ItemTail(elminsterExiled)}.Apply(ctx)
				},
			},
		},
	})
}

// elminsterExiled makes the Faerie Dragons for the exiled card's mana value. An
// on-disk key.
var elminsterExiled = effects.CardTail("then/elminster-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	n := c.ManaValue()
	if n <= 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 blue Faerie Dragon with flying"), N: n}.Apply(ctx)
})

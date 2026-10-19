package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ragavan, Nimble Pilferer — 2/1 Legendary Creature — Monkey Pirate
// for {R}:
//
//	"Whenever Ragavan deals combat damage to a player, create a
//	 Treasure token and exile the top card of that player's library.
//	 Until end of turn, you may cast that card.
//	 Dash {1}{R}"
//
// The archetypal impulse-exile card, and the reason the mechanic
// was worth building: the Treasure is the mana to cast what you
// just stole, so the two halves of the trigger are one play.
//
// Note "you may CAST that card" rather than "play" — a land off
// the top is stranded in exile, which the permission's CastOnly
// flag models. That is not a sandbox shortcut; it's the card.
//
// Dash {1}{R} is the shared Dash constructor (freerunning.go): an alternative
// cost whose entry clause gives haste and returns the creature to its
// owner's hand at the next end step (game/dash.go). No simplification.
func init() {
	Register(Spec{
		OracleID:         "37108cd4-bbab-4ce3-9ed6-f60e8422e703",
		Name:             "Ragavan, Nimble Pilferer",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{1}{R}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID &&
					damagedOpponent(ev, source.Controller, g) != uuid.Nil
			},
			Key: "Ragavan — Treasure, and exile their top card",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Ragavan — Treasure, and exile their top card")
				item.Params.Player = ev.Target
				return item
			},
			Effect: Do(
				// The exile waits for a prompt the Treasure's creation asks (CR 608.2c, #4155).
				CreateToken{Template: TreasureToken(), N: 1},
				Instruction(func(ctx *Context) error {
					return ExileTopWithPermission{
						From:     ctx.Item.Params.Player,
						GrantTo:  ctx.Controller(),
						N:        1,
						CastOnly: true,
					}.Apply(ctx)
				}),
			),
		}},
	})
}

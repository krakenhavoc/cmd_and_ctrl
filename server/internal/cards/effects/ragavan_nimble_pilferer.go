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
// Dash {1}{R} is an alternative cast path (S29) and isn't modelled.
func init() {
	Register(Spec{
		OracleID:     "37108cd4-bbab-4ce3-9ed6-f60e8422e703",
		Name:         "Ragavan, Nimble Pilferer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Dash isn't available — Ragavan can only be cast normally for {R}."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID &&
					damagedOpponent(ev, source.Controller, g) != uuid.Nil
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				victim := ev.Target
				return game.NewTriggeredItem(source, "Ragavan — Treasure, and exile their top card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if err := (CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          1,
						}).Apply(ctx); err != nil {
							return err
						}
						return ExileTopWithPermission{
							From:     victim,
							GrantTo:  item.Controller,
							N:        1,
							CastOnly: true,
						}.Apply(ctx)
					})
			},
		}},
	})
}

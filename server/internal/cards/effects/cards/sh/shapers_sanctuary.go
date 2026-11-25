package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shapers' Sanctuary — Enchantment {G} (Ixalan):
//
//	"Whenever a creature you control becomes the target of a spell or ability
//	 an opponent controls, you may draw a card."
//
// Unsettled Mariner's becomes-the-target watch (CR 115.9, one trigger per
// target chosen), narrowed to creatures on the battlefield under its
// controller's control and to an opponent's spell or ability, with a CR 603.5
// "you may" asked as it resolves.
//
// No simplification.
func init() {
	const label = "Shapers' Sanctuary — you may draw a card"
	effects.Register(effects.Spec{
		OracleID:     "15bd8fa1-0858-4029-9973-e24c86e45af3",
		Name:         "Shapers' Sanctuary",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventBecomesTarget},
				Key:     label,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.StackItemID == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller || ev.CardID == uuid.Nil {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && g.Battlefield.Contains(ev.CardID) && c.IsCreature() && c.Controller == source.Controller
				},
				Effect: effects.Do(effects.DrawCards{N: 1}),
			}, "Shapers' Sanctuary — draw a card?"),
		},
	})
}

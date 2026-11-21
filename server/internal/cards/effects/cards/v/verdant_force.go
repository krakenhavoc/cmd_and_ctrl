package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Verdant Force — Creature — Elemental {5}{G}{G}{G}, 7/7 (Commander Starter
// Decks):
//
//	"At the beginning of each upkeep, create a 1/1 green Saproling creature
//	 token."
//
// Wolverine Riders' upkeep trigger over a Saproling: every upkeep, not just
// its controller's, so a four-player table makes four a round. The row
// watches EventBeginUpkeep with no owner gate.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a21ea22-3cd7-4c11-8895-5943c0d93a0d",
		Name:         "Verdant Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBeginUpkeep
			}, "Verdant Force — create a 1/1 green Saproling", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})),
		},
	})
}

package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// First Response — Enchantment {3}{W} (Commander Legends):
//
//	"At the beginning of each upkeep, if you lost life last turn, create a
//	 1/1 white Soldier creature token."
//
// Feast on the Fallen's "last turn" (the previous turn, whoever's it was; the
// engine stamps the players who lost life as the turn begins), read for the
// controller. An intervening "if" (CR 603.4), checked as each upkeep begins and
// again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fe66bc0-8eda-4625-a068-6a5de826466f",
		Name:         "First Response",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.S82OldB083ILostLifeLastTurn,
				effects.On(game.EventBeginUpkeep, effects.S82OldB083ILostLifeLastTurnWhen,
					"First Response — create a 1/1 white Soldier creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}))),
		},
	})
}

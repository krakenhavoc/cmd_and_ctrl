package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Test of Endurance — Enchantment {2}{W}{W} (Dominaria Remastered, #5948):
//
//	"At the beginning of your upkeep, if you have 50 or more life, you win
//	 the game."
//
// Felidar Sovereign's trigger at 50: an intervening "if" (CR 603.4), checked
// when the upkeep begins and again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffdad8fb-69ee-40d3-b2eb-b0aee0a5becc",
		Name:         "Test of Endurance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.LifeAtLeast(g, source.Controller, 50)
			}, "Test of Endurance — if you have 50 or more life, you win the game", effects.WinIfLifeAtLeast(50)),
		},
	})
}

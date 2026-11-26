package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// World Queller — Creature — Avatar {3}{W}{W}, 4/4 (Zendikar, #5583):
//
//	"At the beginning of your upkeep, you may choose a card type. If you do,
//	 each player sacrifices a permanent of their choice of that type."
//
// The choice is a prompt whose first branch declines. Each player, you
// included, then sacrifices a permanent of the chosen type of their own
// choice, in APNAP order (CR 101.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ac27b74-aa44-412d-8099-ace5009628f0",
		Name:         "World Queller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("World Queller — you may choose a card type; each player sacrifices a permanent of that type",
				effects.OldB497WorldQuellerUpkeep),
		},
	})
}

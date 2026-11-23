package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lore Broker — Creature — Human Rogue {1}{U}, 1/2 (Ravnica: City of Guilds,
// #5600):
//
//	"{T}: Each player draws a card, then discards a card."
//
// Geier Reach Sanitarium's table loot on a tap row: each player, the
// controller first, draws and then chooses their own discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e60a9e8d-0be4-4d08-a03a-af9f9e2be65b",
		Name:         "Lore Broker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Each player draws a card, then discards a card.",
			// ADR 0126 §6: every other player loots too, which no field says.
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Cost:    effects.TapCost(),
			Effect:  effects.EachPlayerDrawsThenDiscardsACard,
		}},
	})
}

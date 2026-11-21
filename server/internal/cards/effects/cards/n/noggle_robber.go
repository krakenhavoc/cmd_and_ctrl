package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noggle Robber — Creature — Noggle Rogue {1}{R/G}{R/G} (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters or dies, create a Treasure token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2f33166-6cfc-4dc5-8edc-088bf40f06b5",
		Name:         "Noggle Robber",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Noggle Robber — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}

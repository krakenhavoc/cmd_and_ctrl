package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dire Fleet Hoarder — Creature — Human Pirate {1}{B}, 2/1:
//
//	"When this creature dies, create a Treasure token."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d48ad95-d428-4e2d-b697-56907c3ce579",
		Name:         "Dire Fleet Hoarder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Dire Fleet Hoarder — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}

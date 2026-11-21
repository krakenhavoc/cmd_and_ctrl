package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Common Crook — Creature — Human Rogue Villain {1}{B}, 2/2 (Marvel Super
// Heroes Commander, #4455): "When this creature dies, create a Treasure
// token." A dies trigger (CR 700.4) that makes the catalog Treasure.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebcbfdba-34c7-40eb-91a9-d306ff5ae8b9",
		Name:         "Common Crook",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Common Crook — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}

package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Easterling Vanguard — Creature — Human Warrior {1}{B}, 2/1 (The Lord of the
// Rings: Tales of Middle-earth):
//
//	"When this creature dies, amass Orcs 1."
//
// Grim Initiate's dies-amass (WhenThisDies), at the Orc subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "149887fa-6b2d-495e-b3e0-b2ac7d3dc838",
		Name:         "Easterling Vanguard",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Easterling Vanguard — amass Orcs 1", effects.Do(effects.Amass{Subtype: "Orc", N: 1})),
		},
	})
}

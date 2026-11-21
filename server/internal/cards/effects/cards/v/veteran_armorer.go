package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Armorer — Creature — Human Soldier {1}{W}, 2/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Other creatures you control get +0/+1."
//
// A layer 7c anthem over the other creatures its controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b1214e9-1581-4008-b3d4-8b7af6b092fd",
		Name:         "Veteran Armorer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 0, 1)},
	})
}

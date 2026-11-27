package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Snowcrusher — Snow Artifact Creature — Phyrexian Juggernaut {6}, 6/5 (Coldsnap, S82 OLD b560):
//
//	"This creature attacks each combat if able.
//	 {1}{S}: This creature gets +1/+0 until end of turn."
//
// {S} is paid with snow mana only (CR 107.4h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f20aa69a-e2eb-4103-9684-12390ef07c6c",
		Name:         "Phyrexian Snowcrusher",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Activated:    []effects.ActivatedAbility{effects.ThisGetsRow("{1}{S}", 1, 0, "Phyrexian Snowcrusher")},
	})
}

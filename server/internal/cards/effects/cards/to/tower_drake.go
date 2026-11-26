package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tower Drake — Creature — Drake {2}{U}, 2/1 (Return to Ravnica, S82 OLD):
//
//	"Flying
//	 {W}: This creature gets +0/+1 until end of turn."
//
// Beacon Hawk's pump row on a blue flyer; the {W} is paid from any source of
// white mana the player has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4fee183e-973e-4be7-bb68-3e1c0c6bfb5b",
		Name:            "Tower Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{W}", 0, 1, "Tower Drake")},
	})
}

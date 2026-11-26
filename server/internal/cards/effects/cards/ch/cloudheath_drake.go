package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudheath Drake — Artifact Creature — Drake {4}{U}, 3/3 (Shards of Alara,
// S82 OLD batch 510, #5588):
//
//	"Flying
//	 {1}{W}: This creature gains vigilance until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "95500513-5519-4131-bb89-42383289f2bd",
		Name:            "Cloudheath Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{1}{W}", "vigilance")},
	})
}

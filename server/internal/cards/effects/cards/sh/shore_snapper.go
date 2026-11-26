package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shore Snapper — Creature — Beast {2}{B}, 2/2 (Shards of Alara, S82 OLD
// batch 510, #5588):
//
//	"{U}: This creature gains islandwalk until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5e2ed91-edd2-47ec-809e-3e4a4245e2ce",
		Name:         "Shore Snapper",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{U}", "islandwalk")},
	})
}

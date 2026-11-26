package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Champion's Drake — Creature — Drake {1}{U}, 1/1 (Rise of the Eldrazi, S82
// OLD batch 482, #5581):
//
//	"Flying
//	 This creature gets +3/+3 as long as you control a creature with three or
//	 more level counters on it."
//
// The Drake counts itself if it were levelled. A layer 7c static of the
// source alone, read live from the battlefield's level counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3aca5718-9151-4a5b-b8a3-bfac7abf65ef",
		Name:            "Champion's Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.OldB482SelfPumpWhileYouControlLevelCreature(3, 3, 3)},
	})
}

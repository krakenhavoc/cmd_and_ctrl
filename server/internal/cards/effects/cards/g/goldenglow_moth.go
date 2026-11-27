package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goldenglow Moth — Creature — Insect {W}, 0/1 (Magic 2011, #5579):
//
//	"Flying
//	 Whenever this creature blocks, you may gain 4 life."
//
// A "you may" trigger on this creature's own block declaration (once for each
// creature it blocks, CR 509.3b); the question is asked as it resolves (CR
// 603.5) and declining gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8dcfa697-5350-43b9-896f-c5eefcc3893a",
		Name:            "Goldenglow Moth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBlock, effects.S82OldB476ThisBlocked, "Goldenglow Moth — you may gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
				"Goldenglow Moth — gain 4 life?"),
		},
	})
}

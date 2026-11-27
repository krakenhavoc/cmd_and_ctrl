package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steam Spitter — Creature — Spider {4}{G}, 1/5 (Coldsnap, S82 OLD b560):
//
//	"Reach
//	 {R}: This creature gets +1/+0 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "76ed34d2-e09e-4b53-98bd-d96b367fd315",
		Name:            "Steam Spitter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{R}", 1, 0, "Steam Spitter")},
	})
}

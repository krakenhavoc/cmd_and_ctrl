package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stomper Cub — {3}{G}{G} Creature — Beast 5/3 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Trample

func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d448c85-9c93-4e9e-8aa0-0035fc94112e",
		Name:            "Stomper Cub",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}

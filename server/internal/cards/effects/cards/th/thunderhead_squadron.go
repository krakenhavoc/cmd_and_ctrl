package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderhead Squadron — Creature — Human Knight {5}{U} 3/4 (March of the
// Machine): "Convoke. Flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d74d6f6a-aff0-4a36-9a99-8b2281f5e29d",
		Name:            "Thunderhead Squadron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         effects.Convoke(),
	})
}

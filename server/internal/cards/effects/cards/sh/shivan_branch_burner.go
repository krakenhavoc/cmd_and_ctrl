package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shivan Branch-Burner — Creature — Dragon {5}{R}{R} 4/4 (March of the
// Machine): "Convoke. Flying, haste."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf7f5e9d-de3d-4a29-af25-f0b56336cecb",
		Name:            "Shivan Branch-Burner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		TapCost:         effects.Convoke(),
	})
}

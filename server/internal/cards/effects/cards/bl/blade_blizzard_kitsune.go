package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade-Blizzard Kitsune — Creature — Fox Ninja {2}{W}, 2/2 (Neon Dynasty): "Ninjutsu {3}{W} / Double strike".
//
// Ninjutsu is the shared keyword (ninjutsu.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84fb1c2f-cf1b-4166-834f-a887aa182d09",
		Name:            "Blade-Blizzard Kitsune",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated:       []effects.ActivatedAbility{effects.Ninjutsu("{3}{W}")},
	})
}

package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dokuchi Shadow-Walker — Creature — Ogre Ninja {4}{B}{B}, 5/5 (Neon
// Dynasty): "Ninjutsu {3}{B}". Nothing else is printed.
//
// Ninjutsu is the shared keyword (ninjutsu.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e79b7bf-8a2d-4585-81d0-755dce78a4ea",
		Name:         "Dokuchi Shadow-Walker",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{3}{B}")},
	})
}

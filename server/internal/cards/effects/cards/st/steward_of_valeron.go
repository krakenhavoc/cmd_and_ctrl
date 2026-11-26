package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steward of Valeron — Creature — Human Druid Knight {G}{W}, 2/2 (Shards of
// Alara, S82 OLD batch 517, #5588):
//
//	"Vigilance
//	 {T}: Add {G}."
//
// Vigilance is the keyword's own; the mana ability is the plain {T}: Add {G}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca5d0a56-69c1-4bdb-95f0-af0dec3604a5",
		Name:            "Steward of Valeron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}

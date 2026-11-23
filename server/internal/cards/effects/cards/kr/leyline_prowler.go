package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leyline Prowler — Creature — Nightmare Beast {1}{B}{G}, 2/3 (War of the Spark, #5520):
//
//	"Deathtouch, lifelink
//	 {T}: Add one mana of any color."
//
// Printed keywords plus Birds of Paradise's mana ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ca3852a-fa37-46ef-a74d-3c7487ae50fc",
		Name:            "Leyline Prowler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
	})
}

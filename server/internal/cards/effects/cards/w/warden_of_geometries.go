package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warden of Geometries — Creature — Eldrazi Drone {4}, 2/3 (Oath of the Gatewatch, #5545):
//
//	"Vigilance
//	 {T}: Add {C}. ({C} represents colorless mana.)"
//
// Vigilance and Eldrazi Temple's {T}: Add {C}. A creature's {T} mana
// ability waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "448da389-81db-4931-9cf7-3fde0ce9392f",
		Name:            "Warden of Geometries",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities:   []effects.ManaAbility{effects.PainlessColorless()},
	})
}

package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akoum Warrior // Akoum Teeth — Modal DFC. Front: Creature — Minotaur
// Warrior {5}{R}, 4/5, Trample. Back: Land, "This land enters tapped.
// {T}: Add {R}." (the back face is in mdfc_lands.go, under
// "<oracle_id>#1").
//
// The front is a vanilla trampler: the keyword rides PrintedKeywords and
// combat does the rest. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "afedce7b-0e18-40ad-a26a-1933fddb560d",
		Name:            "Akoum Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}

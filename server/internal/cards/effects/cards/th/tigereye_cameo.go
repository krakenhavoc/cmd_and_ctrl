package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tigereye Cameo — Artifact {3}:
//
//	"{T}: Add {G} or {W}."
//
// Selesnya Cluestone's mana ability with nothing else: a two-colour rock, the
// colour picked as the mana is added (the auto-tapper books whichever the cast
// needs).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2be289e-e560-405d-9728-d8a4ee9cbf56",
		Name:         "Tigereye Cameo",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
	})
}

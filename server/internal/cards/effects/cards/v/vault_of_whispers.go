package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vault of Whispers — Artifact Land (EDHREC rank 612):
//
//	"{T}: Add {B}."
//
// Seat of the Synod's black twin: a Swamp that is also an artifact.
// The printed type line carries no basic land type, so the engine's
// synthetic land ability gives it nothing and the one fixed {B} is
// declared here. It is an artifact on the battlefield, so it counts
// for Storm-Kiln Artist and every "artifacts you control" clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09496421-74e4-466a-9546-56f2a0c8eef4",
		Name:         "Vault of Whispers",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}

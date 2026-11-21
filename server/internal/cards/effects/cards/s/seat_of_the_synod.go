package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seat of the Synod — Artifact Land (EDHREC rank 317):
//
//	"{T}: Add {U}."
//
// An Island that is also an artifact — the affinity and
// artifact-count decks' free artifact. The type line is printed
// "Artifact Land" with no basic type, so the engine's synthetic
// land ability (basic supertype only) gives it nothing; the one
// fixed {U} is declared here. IsArtifact reads the printed type
// line, so it counts for Storm-Kiln Artist and any "artifacts you
// control" clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39451b4d-cd7a-40da-b457-cb51b609173f",
		Name:         "Seat of the Synod",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}

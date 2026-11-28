package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tolarian Academy — Legendary Land (Urza's Saga, S82 OLD batch 731, #5624):
//
//	"{T}: Add {U} for each artifact you control."
//
// A mana ability (CR 605.1a) whose amount is counted as it resolves, through
// the layers: every artifact its controller controls, whatever else it is
// (an artifact creature or land counts). With none it adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dba4fd31-8931-42dd-bd86-45479c2abf74",
		Name:         "Tolarian Academy",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("U", effects.MatchArtifact),
			Label:          "Add {U} for each artifact you control",
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Igneous Golem — Artifact Creature — Golem {5}, 3/4 (Mirage, #5633):
//
//	"{2}: This creature gains trample until end of turn."
//
// A layer-6 grant pinned to the ability's source through the shared
// thisGainsKeywordUntilEndOfTurn body; trample is read as combat damage
// is assigned (CR 702.19b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c4408eb-b3e2-407f-9018-197787edeae7",
		Name:         "Igneous Golem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"trample"}}},
			Cost:    ManaCost("{2}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("trample", "Igneous Golem — trample until end of turn"),
		}},
	})
}

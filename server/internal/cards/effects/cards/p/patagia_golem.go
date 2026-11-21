package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patagia Golem — Artifact Creature — Golem {4}, 2/3 (Mirage, #5633):
//
//	"{3}: This creature gains flying until end of turn."
//
// A layer-6 grant pinned to the ability's source through the shared
// thisGainsKeywordUntilEndOfTurn body, so the Golem can't be blocked
// except by creatures with flying or reach while it lasts (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85308fb8-e5a1-4ff1-8606-283cf2056895",
		Name:         "Patagia Golem",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"flying"}}},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Patagia Golem — flying until end of turn"),
		}},
	})
}

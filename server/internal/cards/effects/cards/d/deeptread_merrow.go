package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deeptread Merrow — Creature — Merfolk Rogue {1}{U}, 2/1 (Lorwyn, S82 OLD
// b547, #5593):
//
//	"{U}: This creature gains islandwalk until end of turn."
//
// A keyword grant on the source, ending in the cleanup step (CR 514.2).
// No {T}, so it may be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b301b371-e99e-43cc-af7f-dbd7b2abe5f8",
		Name:         "Deeptread Merrow",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: This creature gains islandwalk until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"islandwalk"}}},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("islandwalk", "Deeptread Merrow — islandwalk until end of turn"),
		}},
	})
}

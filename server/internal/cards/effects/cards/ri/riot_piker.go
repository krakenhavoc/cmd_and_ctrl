package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riot Piker — Creature — Goblin Berserker {1}{R}, 2/1:
//
//	"First strike
//	 This creature attacks each combat if able."
//
// First strike is a printed keyword; "attacks each combat if able" is the
// shared attack requirement Juggernaut uses (CR 508.1d), enforced when
// attackers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "864c81e4-d612-4939-b4df-4b125cb8cdf2",
		Name:            "Riot Piker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
	})
}

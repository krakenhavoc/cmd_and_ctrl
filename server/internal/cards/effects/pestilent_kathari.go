package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pestilent Kathari — Creature — Bird Warrior {2}{B}, 1/1 (Conflux,
// #5587):
//
//	"Flying
//	 Deathtouch (Any amount of damage this deals to a creature is enough
//	 to destroy it.)
//	 {2}{R}: This creature gains first strike until end of turn."
//
// Flying and deathtouch are printed keywords (CR 702.2b). The row is a
// layer-6 grant pinned to the Kathari that ends in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "755ab107-0f65-4536-aa4b-8e9d35a9e524",
		Name:            "Pestilent Kathari",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Activated: []ActivatedAbility{
			conB1ThisGainsRow("{2}{R}", "first strike", "Pestilent Kathari", game.AnswerCombatGrant),
		},
	})
}

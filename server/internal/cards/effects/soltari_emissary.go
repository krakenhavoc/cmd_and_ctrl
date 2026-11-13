package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soltari Emissary — Creature — Soltari Soldier {1}{W}, 2/1
// (Tempest, #5628):
//
//	"{W}: This creature gains shadow until end of turn. (It can block or
//	 be blocked by only creatures with shadow.)"
//
// Endling's "{B}: This creature gains <keyword> until end of turn" with
// shadow, through the shared thisGainsKeywordUntilEndOfTurn body: a
// layer-6 grant pinned to the Emissary (CR 613.1f) that ends in the
// cleanup step (CR 514.2). No tap in the cost, so it can be activated
// while summoning sick. Gaining shadow after blockers are declared
// doesn't remove it from combat or undo a block; the block rules read
// shadow only as blocks are declared (CR 702.28b, CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8211eb1-332d-41c0-8086-4a18cd7a5832",
		Name:         "Soltari Emissary",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature gains shadow until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{W}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("shadow", "Soltari Emissary — shadow until end of turn"),
		}},
	})
}

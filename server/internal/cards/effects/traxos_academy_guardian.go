package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traxos, Academy Guardian — Legendary Artifact Creature — Dragon
// Construct {3}{U}, 1/5:
//
//	"This spell costs {2} less to cast if you've cast a noncreature
//	 spell this turn.
//	 Flying, vigilance
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// A self cost reduction (CR 601.2f) reading the per-turn cast tally,
// which is bumped after a cast succeeds, so Traxos itself (a creature
// spell) is never the spell that turned the discount on. Prowess is a
// canonical keyword the engine turns into a trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b9a574f-6d56-421b-8fbd-94da1e6e3ce1",
		Name:            "Traxos, Academy Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", game.KeywordProwess},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you've cast a noncreature spell this turn.",
				rfCreatureFCastNoncreatureSpellThisTurn()),
		},
	})
}

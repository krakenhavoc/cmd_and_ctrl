package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stensia Masquerade — Enchantment {2}{R}:
//
//	"Attacking creatures you control have first strike.
//	 Whenever a Vampire you control deals combat damage to a player, put a
//	 +1/+1 counter on it.
//	 Madness {2}{R}"
//
// First strike is Okoye's attacking-creatures grant over every creature;
// the counter trigger is Rakish Heir's; madness is the engine's (CR
// 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88f5d146-420b-417c-bcf1-fe0e78312a74",
		Name:         "Stensia Masquerade",
		Completeness: CompletenessFull,
		Madness:      "{2}{R}",
		Static: []game.StaticAbility{
			AttackingCreaturesHave("Attacking creatures you control have first strike.", nil, "first strike"),
		},
		Triggered: []game.TriggeredAbility{s82Rpr29VampireCounterTrigger("Stensia Masquerade")},
	})
}

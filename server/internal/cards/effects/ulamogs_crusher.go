package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ulamog's Crusher — Creature — Eldrazi {8}, 8/8:
//
//	"Annihilator 2
//	 This creature attacks each combat if able."
//
// Annihilator 2 is the canonical keyword token, and the engine makes
// its attack trigger (ADR 0113 §2). "Attacks each combat if able" is
// the shared CR 508.1d requirement: a Crusher that can't attack (tapped,
// summoning sick) or whose attack has a cost is not forced to (the
// 2018-12-07 ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "597bce51-af89-4d13-9a97-667b4f4f4694",
		Name:            "Ulamog's Crusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"annihilator 2"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
	})
}

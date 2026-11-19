package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akroan Skyguard — Creature — Human Soldier {1}{W}, 1/1 (Foundations
// Jumpstart, #5929):
//
//	"Flying
//	 Heroic — Whenever you cast a spell that targets this creature, put a
//	 +1/+1 counter on this creature."
//
// Brigone's heroic counter on a flier.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f4e8be6f-2a1f-47b2-a4d8-7e576e25ab60",
		Name:            "Akroan Skyguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			heroic("Akroan Skyguard — a +1/+1 counter on this creature (heroic)", thisGetsAPlusOneCounter),
		},
	})
}

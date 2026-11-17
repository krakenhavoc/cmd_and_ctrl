package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mishra's Juggernaut — Artifact Creature — Juggernaut {5}, 5/3:
//
//	"Trample
//	 This creature attacks each combat if able.
//	 Unearth {5}{R}"
func init() {
	Register(Spec{
		OracleID:        "fa6fa2f5-eb9b-402e-82f4-e0a8d30df245",
		Name:            "Mishra's Juggernaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
		Activated:       []ActivatedAbility{Unearth("{5}{R}")},
	})
}

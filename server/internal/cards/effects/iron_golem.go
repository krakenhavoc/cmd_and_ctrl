package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Golem — Artifact Creature — Golem {4}, 5/3 (AFR):
//
//	"Vigilance
//	 This creature attacks or blocks each combat if able."
func init() {
	Register(Spec{
		OracleID:        "43990801-ed4a-49ff-a96e-03faa302214f",
		Name:            "Iron Golem",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			AttacksEachCombat(),
			BlocksEachCombat(),
		},
	})
}

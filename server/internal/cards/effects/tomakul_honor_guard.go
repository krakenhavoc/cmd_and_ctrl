package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomakul Honor Guard — Creature — Human Soldier {1}{G}, 3/1 (The Brothers' War, #5482):
//
//	"Ward {2}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b68cf4e6-205e-40b6-b28b-f4b7d8af7017",
		Name:         "Tomakul Honor Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Tomakul Honor Guard — ward {2}"),
		},
	})
}

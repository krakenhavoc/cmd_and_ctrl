package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ambush Paratrooper — Creature — Human Soldier {1}{W}, 1/2 (The Brothers' War, #5482):
//
//	"Flash
//	 Flying
//	 {5}: Creatures you control get +1/+1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd583365-27cc-4170-b138-4b3dd0d842b6",
		Name:            "Ambush Paratrooper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Activated:       []effects.ActivatedAbility{effects.Bro01AmbushParatrooperPump()},
	})
}

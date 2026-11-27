package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Cavalry (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Human Soldier {2}{G}, 1/4:
//	"{W}: This creature gains vigilance until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e6e70eb-6638-4aa6-997b-90f5810e12a0",
		Name:         "Llanowar Cavalry",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB510ThisGainsRow("{W}", "vigilance")},
	})
}

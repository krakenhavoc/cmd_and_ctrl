package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sengir Bats — Creature — Bat {1}{B}{B}, 1/2 (Homelands, S82 OLD b822, #5635):
//
//	"Flying
//	Whenever a creature dealt damage by this creature this turn dies, put a +1/+1
//	counter on this creature."
//
// Sengir Vampire's trigger word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ade8f21a-a7bb-4bd9-a7f2-9fa3060ff1e8",
		Name:            "Sengir Bats",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OldB822DamagedCreatureDies("Sengir Bats — put a +1/+1 counter on it", game.CounterPlusOne),
		},
	})
}

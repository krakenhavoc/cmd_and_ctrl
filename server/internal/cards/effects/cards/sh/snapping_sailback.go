package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snapping Sailback — Creature — Dinosaur {4}{G}, 4/4 (Ixalan, #5532):
//
//	"Flash
//	 Enrage — Whenever this creature is dealt damage, put a +1/+1 counter
//	 on it. (It must survive the damage to get the counter.)"
//
// Siegehorn Ceratops's shared enrage trigger with one counter: one trigger
// per damage instance (CR 120.3), and a Sailback that died to the damage
// is gone by resolution, so the counter goes nowhere (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6f7b9ff4-fe79-4d21-9262-47bcd2411259",
		Name:            "Snapping Sailback",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Snapping Sailback — Enrage: put a +1/+1 counter on it", effects.PlusOneCountersOnThis(1)),
		},
	})
}

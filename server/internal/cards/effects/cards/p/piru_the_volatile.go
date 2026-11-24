package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Piru, the Volatile — Legendary Creature — Elder Dragon {2}{R}{R}{W}{W}{B}{B},
// 7/7 (Modern Horizons 2):
//
//	"Flying, lifelink
//	 At the beginning of your upkeep, sacrifice Piru unless you pay {R}{W}{B}.
//	 When Piru dies, it deals 7 damage to each nonlegendary creature."
//
// The upkeep payment is asked of you as the step begins; a decline (or a pool
// that cannot pay) sacrifices Piru, and a Piru that left, or came back as a new
// object, asks nothing (CR 400.7). The death trigger is Piru's own, and the 7
// damage is dealt by Piru as it last existed to every creature that is not
// legendary, all at once, yours included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dcd13eab-7ed7-4e95-bdae-d140c1ff84df",
		Name:            "Piru, the Volatile",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Piru, the Volatile — sacrifice Piru unless you pay {R}{W}{B}", effects.OldB036PiruUpkeep),
			effects.TriggerWithPurpose(
				effects.WhenThisDies("Piru, the Volatile — 7 damage to each nonlegendary creature", effects.OldB036PiruDies),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 7, Partial: true}}),
		},
	})
}

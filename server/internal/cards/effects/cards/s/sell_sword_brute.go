package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sell-Sword Brute — Creature — Human Mercenary {1}{R}, 2/2 (Ravnica: City
// of Guilds, #5600):
//
//	"When this creature dies, it deals 2 damage to you."
//
// Callous Inspector's dies trigger: the damage comes from the Brute as it
// last existed (CR 608.2h), so it is damage and can be prevented.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7549656b-359f-467e-90ab-6b21d93c26b9",
		Name:         "Sell-Sword Brute",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Sell-Sword Brute — it deals 2 damage to you", effects.Do(effects.DamageController{N: 2})),
		},
	})
}

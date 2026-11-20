package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smoldering Efreet — Creature — Efreet Monk {1}{R}, 2/2 (Fate Reforged,
// #5552):
//
//	"When this creature dies, it deals 2 damage to you."
//
// Callous Inspector's dies trigger: the Efreet deals the damage as it last
// existed on the battlefield (CR 608.2h) to its controller, so a prevention
// or a redirect sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "789b6a61-656e-40c0-be78-9c89c4ecfb0b",
		Name:         "Smoldering Efreet",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Smoldering Efreet — it deals 2 damage to you", Do(DamageController{N: 2})),
		},
	})
}

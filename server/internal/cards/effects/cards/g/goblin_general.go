package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin General — {1}{R}{R} Creature — Goblin Warrior 1/1 (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Whenever this creature attacks, Goblin creatures you control get +1/+1
//	 until end of turn."
//
// The Goblins are the ones you control as the trigger resolves (CR 611.2c),
// the General included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fab0d9c-ccf7-4c6a-94b4-959be139dad8",
		Name:         "Goblin General",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Goblin General — Goblin creatures you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Goblin")),
					Power: 1, Toughness: 1, Label: "Goblin General — +1/+1",
				})),
		},
	})
}

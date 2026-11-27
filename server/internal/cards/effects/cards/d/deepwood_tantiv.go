package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deepwood Tantiv — Creature — Beast {4}{G}, 2/4 (Mercadian Masques, #5619):
//
//	"Whenever this creature becomes blocked, you gain 2 life."
//
// Battle-Scarred Goblin's trigger shape: "becomes blocked" is once per
// blocked attacker however many creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b31a9a7-6c97-4e82-bf72-de83034f81a9",
		Name:         "Deepwood Tantiv",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Deepwood Tantiv — you gain 2 life",
				effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}

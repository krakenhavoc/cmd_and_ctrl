package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Setessan Battle Priest — Creature — Human Cleric {1}{W}, 1/3 (Theros, S82 OLD
// b388):
//
//	"Heroic — Whenever you cast a spell that targets this creature, you gain 2
//	 life."
//
// Heroic (CR 207.2c) with a fixed gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2886098-0309-4da5-b2e5-a06af5cfe5b8",
		Name:         "Setessan Battle Priest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Setessan Battle Priest — you gain 2 life (heroic)", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}

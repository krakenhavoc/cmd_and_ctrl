package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gulping Scraptrap — Creature — Phyrexian Horror {4}{B}, 4/4 (Phyrexia: All
// Will Be One, S82):
//
//	"When this creature enters or dies, proliferate."
//
// One printed trigger over two events (Mogg War Marshal's shape), so one
// registered trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9e48750-d930-4095-abf0-dec2ac53be81",
		Name:         "Gulping Scraptrap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Gulping Scraptrap — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blightbelly Rat — Creature — Phyrexian Rat {1}{B}, 2/2 (Phyrexia: All Will
// Be One, S82):
//
//	"Toxic 1
//	 When this creature dies, proliferate."
//
// Toxic rides PrintedKeywords (the engine's damage tail). The dies trigger is
// one registered trigger; the controller chooses what to proliferate as it
// resolves, and the Rat's own poison on a player is among the choices.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0102259-7a2c-470e-b04b-4a1f615d057c",
		Name:            "Blightbelly Rat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Blightbelly Rat — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}

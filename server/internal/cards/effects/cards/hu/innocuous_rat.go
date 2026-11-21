package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Innocuous Rat — Creature — Rat {1}{B}:
//
//	"When this creature dies, manifest dread."
//
// The trigger is an ordinary dies trigger (CR 700.4), so a bounce or
// an exile does not fire it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "114725ef-ac81-4d13-9153-0a303c3ce9c4",
		Name:         "Innocuous Rat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Innocuous Rat — manifest dread", effects.Do(effects.ManifestDread{})),
		},
	})
}

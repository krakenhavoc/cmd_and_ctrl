package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unsettling Twins — Creature — Human {3}{W}:
//
//	"When this creature enters, manifest dread."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0b2ca8c-cb5a-45f8-ae7c-63ad9ebb4afd",
		Name:         "Unsettling Twins",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Self, "Unsettling Twins — manifest dread", effects.Do(effects.ManifestDread{})),
		},
	})
}

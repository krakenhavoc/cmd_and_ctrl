package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bashful Beastie — Creature — Beast {4}{G}:
//
//	"When this creature dies, manifest dread."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7aeabd9-a72c-47ee-b275-b025d7142766",
		Name:         "Bashful Beastie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Bashful Beastie — manifest dread", effects.Do(effects.ManifestDread{})),
		},
	})
}

package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Panharmonicon — "If an artifact or creature entering causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time."
func init() {
	doubler := effects.DoublesEntering(effects.Or(effects.Artifact(), effects.Creature()))
	doubler.Label = "Panharmonicon"
	doubler.RowLabel = "Abilities that trigger when an artifact or creature enters trigger an additional time."
	effects.Register(effects.Spec{
		OracleID:        "76678885-3674-443d-b9a2-2a460cf6aac0",
		Name:            "Panharmonicon",
		Completeness:    effects.CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Panharmonicon — "If an artifact or creature entering causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time."
func init() {
	doubler := DoublesEntering(Or(Artifact(), Creature()))
	doubler.Label = "Panharmonicon"
	Register(Spec{
		OracleID:        "76678885-3674-443d-b9a2-2a460cf6aac0",
		Name:            "Panharmonicon",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}

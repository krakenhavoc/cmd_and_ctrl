package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Isshin, Two Heavens as One — "If a creature attacking causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time."
func init() {
	doubler := DoublesAttacking(Creature())
	doubler.Label = "Isshin, Two Heavens as One"
	Register(Spec{
		OracleID:        "65114758-9a75-43a7-96e8-0aa68faa6b24",
		Name:            "Isshin, Two Heavens as One",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}

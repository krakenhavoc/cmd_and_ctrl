package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isshin, Two Heavens as One — "If a creature attacking causes a triggered
// ability of a permanent you control to trigger, that ability triggers an
// additional time."
func init() {
	doubler := effects.DoublesAttacking(effects.Creature())
	doubler.Label = "Isshin, Two Heavens as One"
	effects.Register(effects.Spec{
		OracleID:        "65114758-9a75-43a7-96e8-0aa68faa6b24",
		Name:            "Isshin, Two Heavens as One",
		Completeness:    effects.CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}

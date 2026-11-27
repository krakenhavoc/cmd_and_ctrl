package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balduvian Barbarians — Creature — Human Barbarian {1}{R}{R}, 3/2 (Eighth Edition, #5607). No rules text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e162cc1-9482-4812-8897-209ec2927ec0",
		Name:         "Balduvian Barbarians",
		Completeness: effects.CompletenessFull,
	})
}

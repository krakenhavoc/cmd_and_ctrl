package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hurr Jackal — Creature — Jackal {R}, 1/1:
//
//	"{T}: Target creature can't be regenerated this turn."
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d17f5afa-a884-4b99-aa9e-89ddb3d43b22",
		Name:         "Hurr Jackal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature can't be regenerated this turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeRegeneratedThisTurn,
		}},
	})
}

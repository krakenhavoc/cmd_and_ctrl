package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furnace Brood — Creature — Elemental {3}{R}, 3/3:
//
//	"{R}: Target creature can't be regenerated this turn."
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc87a268-253b-4377-bdc5-474940de8878",
		Name:         "Furnace Brood",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature can't be regenerated this turn.",
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBeRegeneratedThisTurn,
		}},
	})
}

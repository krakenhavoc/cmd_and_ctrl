package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divine Retribution — Instant {1}{W} (Mirage, #5633):
//
//	"Divine Retribution deals damage to target attacking creature equal to the
//	 number of attacking creatures."
//
// The count includes the target and is taken as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b08b348c-cf74-4aea-ab88-38201e04e42c",
		Name:         "Divine Retribution",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:    effects.OldB794DamageEqualToAttackers,
	})
}

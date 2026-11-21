package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final Strike — Sorcery {2}{B}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Final Strike deals damage to target opponent or planeswalker equal
//	 to the sacrificed creature's power."
//
// The sacrificed creature's last-known power (CR 608.2h), read off the
// payment record (ADR 0113 §1); negative power deals none (CR 107.1b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "4d98aea2-b4ff-4903-ba28-a53fbfaad6b1",
		Name:           "Final Strike",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetOpponentOrPlaneswalker(),
		OnResolve:      effects.DamageEqualToSacrificedPower,
	})
}

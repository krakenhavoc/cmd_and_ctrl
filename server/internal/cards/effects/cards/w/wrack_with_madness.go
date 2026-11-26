package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrack with Madness — Sorcery {3}{R}:
//
//	"Target creature deals damage to itself equal to its power."
//
// Inner Struggle's resolution: the creature is the source of its own damage,
// so deathtouch and lifelink on it apply, and its power is read as the spell
// resolves (CR 608.2b, 120.3).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c93c647e-0b9e-4a50-a519-42862d21a0f1",
		Name:         "Wrack with Madness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Harmful(effects.TargetCreature("target creature")),
		OnResolve:    effects.OldB287TargetDealsDamageToItself,
	})
}

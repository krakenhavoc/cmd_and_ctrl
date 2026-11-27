package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiku's Shadow — Sorcery {B}{B} (Saviors of Kamigawa, S82 OLD b583):
//
//	"Target creature deals damage to itself equal to its power."
//
// Justice Strike's body (B583TargetCreatureDamagesItself): the creature is its
// own damage source, so lifelink, deathtouch and protection read it, and its
// power is read as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce1678a1-441a-4303-a83a-d7d138b366fd",
		Name:         "Kiku's Shadow",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.B583TargetCreatureDamagesItself,
	})
}

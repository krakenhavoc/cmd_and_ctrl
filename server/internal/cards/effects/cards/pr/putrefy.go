package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Putrefy — Instant {1}{B}{G}:
//
//	"Destroy target artifact or creature. It can't be regenerated."
//
// Mortify's Golgari counterpart — artifact instead of enchantment.
// Same regeneration rider, enforced since #667.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b271430-f53d-42d6-a547-2f286dd9bcb6",
		Name:         "Putrefy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}

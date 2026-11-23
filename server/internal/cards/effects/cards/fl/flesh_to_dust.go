package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flesh to Dust — Instant {3}{B}{B} (Magic 2015, #5556):
//
//	"Destroy target creature. It can't be regenerated."
//
// Terminate's body (destroyTheTargetPermanentNoRegen): the no-regeneration
// rider rides the destroy, so a regeneration shield is not spent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f05f3ed-e3b7-4225-b3f2-c12eda335237",
		Name:         "Flesh to Dust",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}

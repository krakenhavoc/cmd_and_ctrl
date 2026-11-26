package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inner Struggle — Instant {3}{R} (Shadows over Innistrad, S82 OLD):
//
//	"Target creature deals damage to itself equal to its power."
//
// Solar Blaze's self-damage for one chosen creature: the creature is the
// source, so deathtouch, lifelink and protection from itself all apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c993ebf8-972f-41c7-9870-d7f8aa9a69d9",
		Name:         "Inner Struggle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Harmful(effects.TargetCreature("target creature")),
		OnResolve:    effects.OldB287TargetDealsDamageToItself,
	})
}

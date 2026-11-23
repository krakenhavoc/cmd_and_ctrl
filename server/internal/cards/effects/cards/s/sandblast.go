package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandblast — Instant {2}{W} (Fate Reforged,
// #5552):
//
//	"Sandblast deals 5 damage to target attacking or blocking creature."
//
// Divine Arrow's shape for 5: the target must be attacking or blocking when
// chosen and again as the spell resolves (CR 608.2b); a creature removed
// from combat in response takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41ed9b02-5926-4af0-bcee-7f41bbb3fdd6",
		Name:         "Sandblast",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.DamageToFirstTargetStep(5),
	})
}

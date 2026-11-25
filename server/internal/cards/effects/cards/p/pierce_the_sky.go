package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pierce the Sky — Instant {1}{G} (Dominaria, S82 OLD, #5530):
//
//	"Pierce the Sky deals 7 damage to target creature with flying."
//
// Flying is read through the layered view when the target is chosen and again
// as the spell resolves (CR 608.2b), so a creature that lost flying in
// response is an illegal target and the spell is countered on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e4504c1-4349-408d-8fdf-0c44efdd3f56",
		Name:         "Pierce the Sky",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 7)),
		OnResolve:    effects.DamageToFirstTargetStep(7),
	})
}

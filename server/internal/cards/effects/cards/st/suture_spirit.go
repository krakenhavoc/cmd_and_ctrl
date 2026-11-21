package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suture Spirit — Creature — Spirit {1}{W}, 1/1 (Eventide, #5590):
//
//	"Flying
//	 {W/B}{W/B}{W/B}: Regenerate target creature."
//
// The shield goes on the target as the ability resolves (CR 701.19a); a
// target gone by then gets nothing (lrwB1RegenerateTargetRow).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b2eba61-3937-4dc2-ad0a-e51fff8ecb80",
		Name:            "Suture Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.LrwB1RegenerateTargetRow("{W/B}{W/B}{W/B}: Regenerate target creature.", effects.ManaCost("{W/B}{W/B}{W/B}"), effects.TargetCreature("target creature")),
		},
	})
}

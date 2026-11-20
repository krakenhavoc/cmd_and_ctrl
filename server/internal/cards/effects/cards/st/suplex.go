package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suplex — Sorcery {1}{R}:
//
//	"Choose one —
//	 • Suplex deals 3 damage to target creature. If that creature would
//	   die this turn, exile it instead.
//	 • Exile target artifact."
//
// Agate Assault's two bullets for 3 damage: the creature is marked
// whether or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d1bbeda-2e81-4aaa-9494-094ec3dd6c3b",
		Name:         "Suplex",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Suplex deals 3 damage to target creature. If that creature would die this turn, exile it instead.",
				effects.TargetCreature("target creature"), effects.DamageModesTargetExileIfItDies(3)), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.ModeDoing("Exile target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()), effects.MurdocksExileTheModesTarget),
		),
	})
}

package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agate Assault — Sorcery {2}{R}:
//
//	"Choose one —
//	 • Agate Assault deals 4 damage to target creature. If that creature
//	   would die this turn, exile it instead.
//	 • Exile target artifact."
//
// The first bullet is Lava Coil's sentence: the replacement is the
// spell's, so the creature is marked whether or not the damage is dealt
// (ADR 0108 §1). The second is Murdock's Crusade's exile bullet.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b89b7d2-c724-4d5d-9f0b-7d3302ad1168",
		Name:         "Agate Assault",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Agate Assault deals 4 damage to target creature. If that creature would die this turn, exile it instead.",
				effects.TargetCreature("target creature"), effects.DamageModesTargetExileIfItDies(4)), effects.ForTargets(effects.DamageToTarget(0, 4))),
			effects.ModeDoing("Exile target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()), effects.MurdocksExileTheModesTarget),
		),
	})
}

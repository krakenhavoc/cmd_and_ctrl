package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pinecone Strike — Instant {1}{R}:
//
//	"Choose one or both —
//	 • Pinecone Strike deals 3 damage to target creature. If that
//	   creature would die this turn, exile it instead.
//	 • Destroy target artifact token."
//
// Each bullet reads its own target group, and the chosen bullets
// resolve in printed order (CR 608.2c). The creature is marked whether
// or not the damage is dealt (ADR 0108 §1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "961e3023-39ea-4141-99b6-738280a2815d",
		Name:         "Pinecone Strike",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Pinecone Strike deals 3 damage to target creature. If that creature would die this turn, exile it instead.",
				effects.TargetCreature("target creature"), effects.DamageModesTargetExileIfItDies(3)), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.ModeDoing("Destroy target artifact token.",
				effects.TargetPermanent("target artifact token", effects.Artifact(), effects.IsTokenPredicate()), effects.DestroyTheModesTarget),
		),
	})
}

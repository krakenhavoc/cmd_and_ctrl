package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Annihilating Glare — Sorcery {B}:
//
//	"As an additional cost to cast this spell, pay {4} or sacrifice an
//	 artifact or creature. Destroy target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2), branches in printed order: branch 0
// is the {4}, which is exactly why the announcement's branch is required
// rather than defaulted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a75087aa-41bd-4e0f-a118-61b5209356f5",
		Name:         "Annihilating Glare",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.ManaAdditionalCost("{4}").Keyed("mana"),
			effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).Keyed("sacrifice"),
		),
		Targets:   effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}

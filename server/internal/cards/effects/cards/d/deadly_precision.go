package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadly Precision — Sorcery {B}:
//
//	"As an additional cost to cast this spell, pay {4} or sacrifice an
//	 artifact or creature. Destroy target creature."
//
// The either/or cost (ADR 0100 §2), branches in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d23ca3d2-0731-4df4-a3c6-69ddf0c58601",
		Name:         "Deadly Precision",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.ManaAdditionalCost("{4}").Keyed("mana"),
			effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).Keyed("sacrifice"),
		),
		Targets:   effects.TargetCreature("target creature"),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}

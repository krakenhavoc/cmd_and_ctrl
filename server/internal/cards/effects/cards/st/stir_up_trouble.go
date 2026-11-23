package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stir Up Trouble — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature or pay {4}. Destroy target creature."
//
// The either/or cost (ADR 0100 §2), branches in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dda607bd-f419-4b7f-b052-a5ce6ce22bfe",
		Name:         "Stir Up Trouble",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).Keyed("sacrifice"),
			effects.ManaAdditionalCost("{4}").Keyed("mana"),
		),
		Targets:   effects.TargetCreature("target creature"),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}

package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silence the Echo — Sorcery {1}{B} (Reality Fracture, tracker #2795):
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 planeswalker or pay {3}. Destroy target creature or planeswalker."
//
// The either/or additional cost (ADR 0100 §2), Annihilating Glare's shape,
// branches in printed order: the sacrifice, then the {3}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43654381-c55c-4d6d-abfd-12932ceb5a4a",
		Name:         "Silence the Echo",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.EitherCost(
			effects.SacrificeCost("a creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())).Keyed("sacrifice"),
			effects.ManaAdditionalCost("{3}").Keyed("mana"),
		),
		Targets:   effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}

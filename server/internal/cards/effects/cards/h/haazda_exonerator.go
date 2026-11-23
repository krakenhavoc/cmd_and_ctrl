package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haazda Exonerator — Creature — Human Cleric {W}, 1/1 (Dissension, #5598):
//
//	"{T}, Sacrifice this creature: Destroy target Aura."
//
// Goblin Digging Team's row with an Aura for the Wall.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6df2e8c9-d87e-4f0d-a3d3-b8eee5cdafd0",
		Name:         "Haazda Exonerator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Destroy target Aura.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target Aura", effects.HasSubtype("Aura")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}

package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nectar Faerie — Creature — Faerie Wizard {1}{B}, 1/1 (Lorwyn):
//
//	"Flying
//	 {B}, {T}: Target Faerie or Elf gains lifelink until end of turn."
//
// The target is any Faerie or Elf permanent, the Nectar Faerie included (it is
// a Faerie). Flying is the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6d0a2e5f-af10-4754-81d7-b7afa6282d7d",
		Name:            "Nectar Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}: Target Faerie or Elf gains lifelink until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target Faerie or Elf", effects.Or(effects.HasSubtype("Faerie"), effects.HasSubtype("Elf"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Nectar Faerie — lifelink"}.Apply(ctx)
			},
		}},
	})
}

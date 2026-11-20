package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Safekeeper — Creature — Human Wizard {G}, 1/1:
//
//	"Sacrifice a land: Target creature you control gains shroud until
//	 end of turn."
//
// An activated ability with a sacrifice cost (Zuran Orb's shape,
// b08SacrificeALand) and a clause that targets one of the
// controller's own creatures. The land is paid at announce, so
// destroying the Safekeeper in response does not return it. Shroud is
// the canonical keyword the targeting choke point already honours, so
// the creature can no longer be targeted by anything, its controller's
// own spells and abilities included, until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bddf8f4a-3149-4dd6-a9e5-7747e7e45a1c",
		Name:         "Sylvan Safekeeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: Target creature you control gains shroud until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "shroud")),
			Cost:    effects.B08SacrificeALand(),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"shroud"},
					Label:    "Sylvan Safekeeper — gains shroud",
				}.Apply(ctx)
			},
		}},
	})
}

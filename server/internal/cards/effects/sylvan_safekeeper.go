package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bddf8f4a-3149-4dd6-a9e5-7747e7e45a1c",
		Name:         "Sylvan Safekeeper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a land: Target creature you control gains shroud until end of turn.",
			Cost:    b08SacrificeALand(),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"shroud"},
					Label:    "Sylvan Safekeeper — gains shroud",
				}.Apply(ctx)
			},
		}},
	})
}

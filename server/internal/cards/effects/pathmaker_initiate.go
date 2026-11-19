package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pathmaker Initiate — Creature — Human Wizard {1}{R}, 2/1 (Foundations
// Jumpstart):
//
//	"{T}: Target creature with power 2 or less can't be blocked this turn."
//
// The target is legal only while its power is 2 or less, on cast and again
// on resolution (CR 608.2b); the restriction is the one Mesmerize grants.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0247a19-9c8f-422c-88cf-9f298c3a8c96",
		Name:         "Pathmaker Initiate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    TapCost(),
			Targets: TargetPermanent("target creature with power 2 or less", And(Creature(), PowerLE(2))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Pathmaker Initiate — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}

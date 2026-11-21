package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c0247a19-9c8f-422c-88cf-9f298c3a8c96",
		Name:         "Pathmaker Initiate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target creature with power 2 or less", effects.And(effects.Creature(), effects.PowerLE(2))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Pathmaker Initiate — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}

package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Access Tunnel — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}: Target creature with power 3 or less can't be blocked
//	 this turn."
//
// Rogue's Passage with a power cap and a cheaper activation. The same
// shape as that card (RestrictUntilEOT, CantBeBlocked, the CR 608.2b
// "skip a target that left" loop) with the power check riding
// PowerLE(3) on the target clause instead of an unrestricted target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed9cc560-f30b-4b60-a094-ccf93ed656a7",
		Name:         "Access Tunnel",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature with power 3 or less can't be blocked this turn",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				// CR 608.2b: a target that left in response is
				// skipped rather than errored.
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Access Tunnel — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

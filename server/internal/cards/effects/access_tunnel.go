package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ed9cc560-f30b-4b60-a094-ccf93ed656a7",
		Name:         "Access Tunnel",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Target creature with power 3 or less can't be blocked this turn",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature with power 3 or less", PowerLE(3)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// CR 608.2b: a target that left in response is
				// skipped rather than errored.
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
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

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minamo Sightbender — Creature — Human Wizard {1}{U}, 1/2:
//
//	"{X}, {T}: Target creature with power X or less can't be blocked
//	 this turn."
//
// ADR 0109 §9 (#1842): the power bound reads the X announced with the
// activation (CR 602.2b), before the target is chosen, and is
// re-checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4f1d8c8-29e5-43fe-8281-7d42d0173dc5",
		Name:         "Minamo Sightbender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{X}, {T}: Target creature with power X or less can't be blocked this turn.",
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Targets: TargetCreature("target creature with power X or less").WithPowerAtMostX(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Minamo Sightbender — can't be blocked",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}

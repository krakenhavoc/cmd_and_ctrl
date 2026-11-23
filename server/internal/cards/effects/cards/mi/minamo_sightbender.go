package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c4f1d8c8-29e5-43fe-8281-7d42d0173dc5",
		Name:         "Minamo Sightbender",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Target creature with power X or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with power X or less").WithPowerAtMostX(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.RestrictUntilEOT{
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

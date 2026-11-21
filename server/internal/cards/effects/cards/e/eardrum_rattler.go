package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eardrum Rattler — Creature — Human Bard {1}{R}, 2/2:
//
//	"{1}, {T}: Another target creature you control with power 2 or less
//	 can't be blocked this turn."
//
// Access Tunnel's shape with the "another ... you control" clause
// (Another sets ExcludeSource, so the Rattler cannot target itself).
// The power limit is checked at announce and again at resolution
// (CR 608.2b), so a creature that grew in response is skipped. The tap
// symbol makes it summoning-sick like any creature's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25151d18-8e73-4b0d-a8d2-e5fa291e5729",
		Name:         "Eardrum Rattler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Another target creature you control with power 2 or less can't be blocked this turn",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature you control with power 2 or less",
				effects.YouControl(), effects.PowerLE(2))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Eardrum Rattler — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

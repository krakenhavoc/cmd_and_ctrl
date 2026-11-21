package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suspicious Bookcase — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "ca611c90-7753-43dd-af91-81694d004eeb", Name: "Suspicious Bookcase", Completeness: effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{Label: "{3}, {T}: Target creature can't be blocked this turn.", Cost: effects.Plus(effects.ManaCost("{3}"), effects.TapCost()), Targets: effects.TargetCreature("target creature"), Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return (effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Suspicious Bookcase — can't be blocked"}).Apply(ctx)
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suspicious Bookcase — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "ca611c90-7753-43dd-af91-81694d004eeb", Name: "Suspicious Bookcase", Completeness: CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{Label: "{3}, {T}: Target creature can't be blocked this turn.", Cost: Plus(ManaCost("{3}"), TapCost()), Targets: TargetCreature("target creature"), Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return (RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Suspicious Bookcase — can't be blocked"}).Apply(ctx)
			},
		}},
	})
}

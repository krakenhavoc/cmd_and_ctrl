package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fertilid — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "21f1c6d7-8289-44b2-b88f-c09e202be200", Name: "Fertilid", Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(game.CounterPlusOne, 2, "Fertilid — enters with two +1/+1 counters")},
		Activated: []ActivatedAbility{{Label: "{1}{G}, Remove a +1/+1 counter from this creature: Target player searches their library for a basic land card, puts it onto the battlefield tapped, then shuffles.", Cost: Plus(ManaCost("{1}{G}"), RemoveCountersFromThis(game.CounterPlusOne, 1)), Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return SearchLibrary{Player: ts[0].ID, Predicate: IsBasicLand, Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, TappedOnEntry: true, Reason: "Fertilid — a basic land, tapped"}.Apply(ctx)
			},
		}},
	})
}

package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fertilid — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "21f1c6d7-8289-44b2-b88f-c09e202be200", Name: "Fertilid", Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Fertilid — enters with two +1/+1 counters")},
		Activated: []effects.ActivatedAbility{{Label: "{1}{G}, Remove a +1/+1 counter from this creature: Target player searches their library for a basic land card, puts it onto the battlefield tapped, then shuffles.", Cost: effects.Plus(effects.ManaCost("{1}{G}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)), Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return effects.SearchLibrary{Player: ts[0].ID, Predicate: effects.IsBasicLand, Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, TappedOnEntry: true, Reason: "Fertilid — a basic land, tapped"}.Apply(ctx)
			},
		}},
	})
}

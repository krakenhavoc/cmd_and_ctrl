package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surge Node — Artifact {1} (New Phyrexia):
//
//	"This artifact enters with six charge counters on it.
//	 {1}, {T}, Remove a charge counter from this artifact: Put a charge
//	 counter on target artifact."
//
// The counters come on as it enters (a self-replacement, CR 614.1c). The
// counter is removed as a cost, and any artifact is a legal target,
// Surge Node itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b30f043-aa79-4262-b13f-6cd794b1920b",
		Name:         "Surge Node",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterCharge, 6, "Surge Node: enters with six charge counters on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Remove a charge counter from this artifact: Put a charge counter on target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 1)),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

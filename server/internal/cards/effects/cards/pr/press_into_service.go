package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Press into Service — Sorcery {4}{R} (Oath of the Gatewatch, #5545):
//
//	"Support 2. (Put a +1/+1 counter on each of up to two target creatures.)
//	 Gain control of target creature until end of turn. Untap that creature. It
//	 gains haste until end of turn."
//
// Two target clauses: the support creatures (any, up to two), then the
// creature to steal. The two are read from their own slots (Bond of
// Passion's shape), so the stolen creature need not get a counter. A clause
// with no legal target does nothing and the other still happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a729a684-9171-4ca9-83ad-1d199afad7bb",
		Name:         "Press into Service",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.SupportTargets(2, false),
			effects.TargetCreature("target creature"),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := effects.NothingIfGone(ctx.Game.AddCounterByForEffect(ctx.Controller(), t.ID, game.CounterPlusOne, 1)); err != nil {
					return err
				}
			}
			t, ok := ctx.ClauseTarget(1)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.GainControl{
				Target:   t.ID,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Press into Service — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"},
				Label: "Press into Service — haste until end of turn"}.Apply(ctx)
		},
	})
}

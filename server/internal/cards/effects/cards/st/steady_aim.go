package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steady Aim — Instant {1}{G} (War of the Spark, #5520):
//
//	"Untap target creature. It gets +1/+4 and gains reach until end of turn."
//
// The untap, then one continuous effect (CR 611.2c): +1/+4 and reach together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5fe1fc3-d9fe-4b1b-85d3-7d8ecb84a829",
		Name:         "Steady Aim",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.ScopedEffectFor{
				Target:   target,
				Mods:     []game.Mod{game.ModifyPTMod(1, 4), game.AddKeywordsMod("reach")},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Steady Aim — +1/+4 and reach until end of turn",
			}.Apply(ctx)
		},
	})
}

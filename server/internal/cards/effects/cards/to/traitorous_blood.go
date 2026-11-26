package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traitorous Blood — Sorcery {1}{R}{R} (Innistrad, S82 older sets batch 449,
// #5572):
//
//	"Gain control of target creature until end of turn. Untap it. It gains
//	 trample and haste until end of turn."
//
// Act of Treason with trample as well: three primitives in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f1a9248-8e7b-46a8-942b-ca18c7da8c4a",
		Name:         "Traitorous Blood",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Traitorous Blood — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"trample", "haste"},
				Label:    "Traitorous Blood — trample and haste until end of turn",
			}.Apply(ctx)
		},
	})
}

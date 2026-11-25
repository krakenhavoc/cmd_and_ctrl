package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Portent of Betrayal — Sorcery {3}{R} (Commander Legends):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. Scry 1."
//
// Act of Treason's body with Pack's Betrayal's trailing scry, which happens
// whether or not the theft did (the target may be gone).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15d794f4-7789-40e3-9ee0-b09250c0516d",
		Name:         "Portent of Betrayal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Scry: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				target := item.Targets[0].ID
				if err := (effects.GainControl{
					Target:   target,
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Portent of Betrayal — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   target,
					Keywords: []string{"haste"},
					Label:    "Portent of Betrayal — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}

package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pack's Betrayal — Sorcery {2}{R} (Innistrad: Midnight Hunt, #5497):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. If you control a Wolf or Werewolf,
//	 scry 2."
//
// Act of Treason's body with a trailing rider. "If you control a Wolf or
// Werewolf" is read as the rider happens, after the theft, so the stolen
// creature counts if it is one (CR 608.2c). It reads effective subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b7abe71-0be3-4dd4-880e-88b65b47ea4d",
		Name:         "Pack's Betrayal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Scry: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Pack's Betrayal — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Pack's Betrayal — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if effects.S82MidControlsWolfOrWerewolf(ctx) {
				return effects.Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}

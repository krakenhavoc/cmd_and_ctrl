package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spreading Insurrection — Sorcery {4}{R} (Modern Horizons 2, #5501):
//
//	"Gain control of target creature you don't control until end of turn.
//	 Untap that creature. It gains haste until end of turn.
//	 Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn. You may choose new targets for the copies.)"
//
// Threaten's body (Bloody Betrayal's, without the Blood) with storm.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc59977f-b3a7-4bc1-9f20-0a1e36d418ae",
		Name:         "Spreading Insurrection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Spreading Insurrection — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Spreading Insurrection — haste until end of turn",
			}.Apply(ctx)
		},
	})
}

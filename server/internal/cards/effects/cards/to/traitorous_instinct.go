package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traitorous Instinct — Sorcery {3}{R} (Return to Ravnica, S82 OLD):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 Until end of turn, it gets +2/+0 and gains haste."
//
// Bloody Betrayal's Act of Treason with a +2/+0 and no Blood token. If the
// target is illegal on resolution the whole spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92d2df7e-9f5c-4c28-a891-b69010bc0001",
		Name:         "Traitorous Instinct",
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
				Label:    "Traitorous Instinct — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 2, Label: "Traitorous Instinct — +2/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Traitorous Instinct — haste until end of turn",
			}.Apply(ctx)
		},
	})
}

package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unwilling Recruit — Sorcery {X}{R}{R}{R}:
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gets +X/+0 and gains haste until end of turn."
//
// Bloody Betrayal's threaten body with X read as the spell resolves (the X
// the spell was cast with, CR 107.3). Control, untap, the pump and haste are
// all pinned to the creature as it was when the spell resolved.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "03a0f7ac-de8a-43cf-be15-07d31e9d6188",
		Name:         "Unwilling Recruit",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.GainControl{Target: t.ID, Duration: effects.DurationUntilEndOfTurn(ctx),
				Label: "Unwilling Recruit — gain control until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Unwilling Recruit — +X/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"},
				Label: "Unwilling Recruit — haste until end of turn"}.Apply(ctx)
		},
	})
}

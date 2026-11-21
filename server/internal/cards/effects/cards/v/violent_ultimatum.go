package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Violent Ultimatum — Sorcery {B}{B}{R}{R}{R}{G}{G} (Shards of Alara,
// #5588):
//
//	"Destroy three target permanents."
//
// Vindicate three times over: one clause of exactly three different
// permanents, each re-checked as the spell resolves, so one that left or
// became illegal is skipped and the others are still destroyed (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a04cfcc5-7bb4-4352-ad8d-9d625585e656",
		Name:         "Violent Ultimatum",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("three target permanents").WithCount(3, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

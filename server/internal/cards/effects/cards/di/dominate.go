package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dominate — Instant {X}{1}{U}{U} (Commander 2015, S82 OLD #5546):
//
//	"Gain control of target creature with mana value X or less."
//
// The target's mana value is bound to the announced X
// (WithManaValueAtMostX), judged at announce (CR 601.2c) and again at
// resolution (CR 608.2b), so a creature whose mana value has grown past X is
// an illegal target and nothing is taken. A permanent with {X} in its cost
// counts that X as 0 (CR 202.3e). The control change has no stated duration,
// so it lasts indefinitely (CR 611.2a): Entrancing Melody's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77db8db4-9fe4-478e-b919-5348b0b87b0d",
		Name:         "Dominate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with mana value X or less").WithManaValueAtMostX(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Dominate — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

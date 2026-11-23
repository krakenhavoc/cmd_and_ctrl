package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sparkmage's Gambit — Sorcery {1}{R} (Double Masters 2022, #5948):
//
//	"Sparkmage's Gambit deals 1 damage to each of up to two target
//	 creatures. Those creatures can't block this turn."
//
// Each target that is still legal at resolution (CR 608.2b) is dealt 1 damage
// by the spell and then barred from blocking until end of turn; one that
// became illegal is skipped. The restriction is applied even when the damage
// was prevented, since the second sentence does not depend on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8df3d1f-130e-4391-9bec-3f2d345a13fb",
		Name:         "Sparkmage's Gambit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if err := ctx.Game.DamageInstanceForEffect(func() error {
				for _, t := range legal {
					if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
			for _, t := range legal {
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Sparkmage's Gambit — can't block this turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})),
	})
}

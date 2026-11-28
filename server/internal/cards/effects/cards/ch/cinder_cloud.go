package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinder Cloud — Instant {3}{R}{R} (Mirage):
//
//	"Destroy target creature. If a white creature dies this way, Cinder Cloud
//	 deals damage to that creature's controller equal to the creature's power."
//
// Whether the creature was white, its power and its controller are read before
// it is destroyed (CR 608.2h); the damage is dealt only if it actually went to
// a graveyard (an indestructible or regenerated creature does not die).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f64ea1a7-0e10-4549-8056-25ebc2ce0175",
		Name:         "Cinder Cloud",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			white, controller, power := victim.HasColorWord("W"), victim.Controller, victim.CurrentPower()
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if !white {
				return nil
			}
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: power}.Apply(ctx)
		},
	})
}

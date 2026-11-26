package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pistus Strike — Instant {2}{G} (Mirrodin Besieged, S82 OLD b467, #5576):
//
//	"Destroy target creature with flying. Its controller gets a poison
//	 counter."
//
// Certain Death's shape: the controller is read before the destroy (CR 608.2h),
// and the poison counter (CR 122.1f) goes to them whether or not the creature
// was destroyed (a regeneration shield or indestructible does not stop it).
// "With flying" is the keyword the creature has now, through the layers, as
// the spell is cast and again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "138f9daa-19d4-4290-94a7-09e00dbcebb0",
		Name:         "Pistus Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := c.Controller
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.AddPlayerCounterByForEffect(ctx.Controller(), controller, game.CounterPoison, 1)
		},
	})
}

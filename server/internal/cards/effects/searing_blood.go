package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Searing Blood — Instant {R}{R} (Avatar: The Last Airbender Eternal, #5075):
//
//	"Searing Blood deals 2 damage to target creature. When that creature
//	 dies this turn, Searing Blood deals 3 damage to the creature's
//	 controller."
//
// The second sentence is a delayed triggered ability the spell creates as
// it resolves (CR 603.7a), bound to the creature as the object it is now
// (CR 400.7): a creature that blinks is a new object and its death does
// nothing. It does not matter what kills the creature this turn (the
// ruling). "The creature's controller" is whoever controlled it as it died
// (the ruling), read from its last-known record (CR 608.2h), and the
// damage comes from Searing Blood. The trigger is registered after the
// damage has been dealt (CR 608.2c), and the creature dies to it only
// after the spell has finished resolving (CR 704.3), so a 2-toughness
// creature fires it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "902b82fd-bb18-4833-b427-af8c9751f870",
		Name:         "Searing Blood",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || !onBattlefield(ctx.Game, t.ID) {
					return nil
				}
				return whenThatCreatureDiesThisTurn(ctx, t.ID, "Searing Blood — 3 damage to the creature's controller", searingBloodDiesBody)
			}),
		),
	})
}

// searingBloodDiesBody is "Searing Blood deals 3 damage to the creature's
// controller": the controller of the creature as it last existed on the
// battlefield (CR 608.2h). An on-disk key.
var searingBloodDiesBody = game.SimpleDelayedBody("searing-blood/3-to-its-controller", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	info, ok := g.PermanentForEffect(obj)
	if !ok || info.Controller == uuid.Nil {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: info.Controller, Amount: 3}.Apply(ctx)
})

package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Virulent Wound — Instant {B} (Mirrodin Besieged, S82 OLD batch 464,
// #5576):
//
//	"Put a -1/-1 counter on target creature. When that creature dies this
//	 turn, its controller gets a poison counter."
//
// Searing Blood's shape: the counter first, then a delayed trigger naming
// that object (CR 603.7), which gives a poison counter to the player who
// controlled the creature as it last existed on the battlefield (CR 608.2h),
// the spell's controller being the one who puts it there (CR 122.1f). A
// creature that was already gone creates no trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a555c10-4dc2-455d-87b4-c7fdbf3d29a6",
		Name:         "Virulent Wound",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || !ctx.IsTargetLegal(t) {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
					return nil
				}
				return effects.WhenThatCreatureDiesThisTurn(ctx, t.ID, "Virulent Wound — its controller gets a poison counter", virulentWoundDiesBody)
			}),
		),
	})
}

// virulentWoundDiesBody is "its controller gets a poison counter". An
// on-disk key: never renamed, never reused.
var virulentWoundDiesBody = game.SimpleDelayedBody("s82-old-b464/virulent-wound-poison", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	obj, ok := effects.PayloadObject(ctx)
	if !ok {
		return nil
	}
	info, ok := g.PermanentForEffect(obj)
	if !ok || info.Controller == uuid.Nil {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.Controller, info.Controller, game.CounterPoison, 1)
})

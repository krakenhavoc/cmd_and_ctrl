package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Knockout Blow — Instant {2}{W} (Streets of New Capenna, #5491):
//
//	"This spell costs {2} less to cast if it targets a red creature.
//	 Knockout Blow deals 4 damage to target attacking or blocking creature
//	 and you gain 2 life."
//
// The reduction reads the announced target (CostsLessIfItTargets). The
// life is gained whether or not the damage lands, as long as the spell
// resolves, i.e. its target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7068637c-b025-41a5-be31-df1850f33b68",
		Name:         "Knockout Blow",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 4)), LifeGain: 2},
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a red creature.",
				And(Creature(), OfColor("R"))),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: 4}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}

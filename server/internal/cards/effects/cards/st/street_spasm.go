package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Street Spasm — Instant {X}{R} (Ravnica: Clue Edition):
//
//	"Street Spasm deals X damage to target creature without flying you
//	 don't control.
//	 Overload {X}{X}{R}{R} (You may cast this spell for its overload cost.
//	 If you do, change "target" in its text to "each.")"
//
// Electrickery's shape with an X amount: the targeted cast and the
// overloaded sweep read one predicate, a creature without flying (through
// the layered view) that its caster doesn't control, so the overload is
// one-sided. X is announced once and paid twice in the overload cost
// (CR 107.3a).
//
// No simplification.
func init() {
	match := effects.And(effects.Creature(), effects.OpponentControls(), effects.WithoutKeyword("flying"))
	effects.Register(effects.Spec{
		OracleID:     "95385d84-550c-4d6c-a889-62bdbc1d518d",
		Name:         "Street Spasm",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature without flying you don't control", effects.OpponentControls(), effects.WithoutKeyword("flying")),
		Purpose:      effects.ForTargets(effects.DamageXToTarget(0)),
		AlternativeCosts: []game.AlternativeCost{
			effects.CostWithPurpose(effects.Overload("{X}{X}{R}{R}"), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true, Partial: true}}),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.PaidAltCost("overload") {
				return effects.DamageEachMatching(ctx, match, ctx.X())
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: ctx.X()}.Apply(ctx)
		},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bond of Passion — Sorcery {4}{R}{R} (War of the Spark, #5520):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. Bond of Passion deals 2 damage to any
//	 other target."
//
// Two target clauses; the second is "any other target" and so must differ from
// the creature (CR 601.2c). Each clause is judged on its own as the spell
// resolves: when only one is still legal, that half still happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "826f9f42-f2ff-4b9e-be79-a9de6b5b7b38",
		Name:         "Bond of Passion",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(1, 2)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(anyOtherTarget()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.GainControl{
					Target:   t.ID,
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Bond of Passion — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"haste"},
					Label:    "Bond of Passion — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok {
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}

// anyOtherTarget is "any other target": the any-target clause, worded for the
// picker. Distinct makes it differ from the creature picked first.
func anyOtherTarget() *game.TargetSpec {
	spec := effects.TargetAny()
	spec.Label = "any other target"
	return spec
}

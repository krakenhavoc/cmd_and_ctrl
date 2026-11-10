package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ram Through — Instant {1}{G}:
//
//	"Target creature you control deals damage equal to its power to
//	 target creature you don't control. If the creature you control has
//	 trample, excess damage is dealt to that creature's controller
//	 instead."
//
// One-sided, like Bite Down: both targets must still be legal as the spell
// resolves or the creature deals no damage. Excess is worked out the way
// trample does (CR 702.19b): lethal damage is toughness minus the damage
// already marked, or 1 when the creature you control has deathtouch
// (CR 702.2c), and prevention or replacement effects are ignored when
// deciding it. The creature and its controller are dealt their damage
// simultaneously, as one damage instance, by the creature you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b1befff-48fa-47a3-832e-bdaf43495c02",
		Name:         "Ram Through",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			src, ok := ctx.Game.LookupCardForEffect(mine.ID)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(theirs.ID)
			if !ok {
				return nil
			}
			power := max(src.CurrentPower(), 0)
			toCreature, toController := power, 0
			if game.HasKeyword(&src, "trample") {
				lethal := max(victim.CurrentToughness()-victim.DamageMarked, 0)
				if game.HasKeyword(&src, "deathtouch") && lethal > 1 {
					lethal = 1
				}
				toCreature = min(power, lethal)
				toController = power - toCreature
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				if toCreature > 0 {
					if err := (DealDamage{Source: mine.ID, Target: theirs.ID, Amount: toCreature}).Apply(ctx); err != nil {
						return err
					}
				}
				if toController > 0 {
					return DealDamage{Source: mine.ID, Target: victim.Controller, Amount: toController}.Apply(ctx)
				}
				return nil
			})
		},
	})
}

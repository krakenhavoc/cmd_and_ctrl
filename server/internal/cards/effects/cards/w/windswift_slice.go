package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windswift Slice — Instant {2}{G} (Tales of Middle-earth Commander):
//
//	"Target creature you control deals damage equal to its power to target
//	 creature you don't control. Create a number of 1/1 green Elf Warrior
//	 creature tokens equal to the amount of excess damage dealt this way."
//
// Bite Down's one-sided bite (slot 0 deals, slot 1 takes, nothing is dealt
// unless both picks are still legal, CR 608.2b) with the excess read as
// Bolg of the North reads it (CR 120.4a): the damage that landed minus the
// damage the creature still needed to die, floored at zero and read after
// the damage so a prevention shield reduces it along with the damage. The
// biter is the damage's source, not the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39ab72e2-04f2-41cc-ae33-4cc564b1542b",
		Name:         "Windswift Slice",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			b, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok {
				return nil
			}
			v, ok := ctx.Game.LookupCardForEffect(victim.ID)
			if !ok {
				return nil
			}
			amount := b.CurrentPower()
			if amount <= 0 {
				return nil
			}
			lethal := max(v.CurrentToughness()-v.DamageMarked, 0)
			cursor := effects.B25LastEventSeq(ctx.Game)
			if err := (effects.DealDamage{Source: biter.ID, Target: victim.ID, Amount: amount}).Apply(ctx); err != nil {
				return err
			}
			excess := effects.B27DamageDealtToAfter(ctx.Game, biter.ID, victim.ID, cursor) - lethal
			if excess <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: excess}.Apply(ctx)
		},
	})
}

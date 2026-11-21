package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinning Wheel Kick — Sorcery {X}{X}{G}{G} (Kamigawa: Neon Dynasty,
// S82 #5493):
//
//	"Target creature you control deals damage equal to its power to each
//	 of X target creatures and/or planeswalkers."
//
// Bite Down's one-sided fight against X targets: two clauses (CR 601.2c),
// the second one's count is the announced X (CountFromX, exactly X, not
// "up to"). Slot 0 is the damage SOURCE, whose power is read as the spell
// resolves and whose deathtouch and lifelink the damage carries; if it has
// become illegal nothing is dealt (CR 608.2b), and each recipient that is
// still legal takes the damage, all at once as one instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed37ce4e-d84e-442a-ac4e-0512f358c138",
		Name:         "Spinning Wheel Kick",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.NeoB15XTargets(),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.X() <= 0 {
				return nil // X is the number of targets: with X = 0 there are none
			}
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok {
				return nil
			}
			power := c.CurrentPower()
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, t := range ctx.ClauseTargets(1) {
					if !ctx.IsTargetLegal(t) {
						continue
					}
					if err := (effects.DealDamage{Source: biter.ID, Target: t.ID, Amount: power}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}

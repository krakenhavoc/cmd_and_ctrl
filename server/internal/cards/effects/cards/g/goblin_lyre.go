package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Lyre — Artifact {3} (Ice Age, S82 OLD batch 840, #5636):
//
//	"Sacrifice this artifact: Flip a coin. If you win the flip, this artifact deals
//	 damage to target opponent or planeswalker equal to the number of creatures you
//	 control. If you lose the flip, this artifact deals damage to you equal to the
//	 number of creatures that opponent or that planeswalker's controller controls."
//
// Mogg Assassin's coin flip: the sacrifice is the cost, the target is chosen as the
// ability goes on the stack, and the creatures are counted when the flip is
// decided (CR 608.2h). Won, the artifact (by last-known information, it is gone)
// deals the damage to the target; lost, it deals damage to its controller equal to
// the creatures controlled by the targeted opponent, or by the controller of the
// targeted planeswalker. A target that is no longer legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0709e9c8-dbe7-4806-ae75-d2e2d745d6a4",
		Name:         "Goblin Lyre",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Sacrifice this artifact: Flip a coin. If you win the flip, this artifact deals damage to target opponent or planeswalker " +
				"equal to the number of creatures you control. If you lose the flip, this artifact deals damage to you equal to the number " +
				"of creatures that opponent or that planeswalker's controller controls.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetOpponentOrPlaneswalker(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(effects.NewContext(g, item).LegalTargets()) == 0 {
					return nil
				}
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Goblin Lyre — call the coin flip",
					ThenKey:  goblinLyreFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

var goblinLyreFlipped = effects.CoinFlipThen("coin-flip/goblin-lyre", func(ctx *effects.Context, r game.CoinFlipped) error {
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	t := targets[0]
	you := ctx.Controller()
	if effects.WonFirstFlip(r) {
		n := len(effects.PermanentsControlledByMatching(ctx.Game, you, effects.Creature()))
		if n == 0 {
			return nil
		}
		return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
	}
	opponent := t.ID
	if t.Kind == game.TargetCard {
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		opponent = c.Controller
	}
	if opponent == uuid.Nil {
		return nil
	}
	n := len(effects.PermanentsControlledByMatching(ctx.Game, opponent, effects.Creature()))
	if n == 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: you, Amount: n}.Apply(ctx)
})

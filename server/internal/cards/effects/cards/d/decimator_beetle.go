package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Decimator Beetle — Creature — Insect {3}{B}{G}, 4/5 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"When this creature enters, put a -1/-1 counter on target creature you
//	 control.
//	 Whenever this creature attacks, remove a -1/-1 counter from target creature
//	 you control and put a -1/-1 counter on up to one target creature defending
//	 player controls."
//
// The attack trigger has two targets (CR 115.1): a creature you control and up
// to one creature the defending player controls, read off this attack. Each
// part is judged on its own as the ability resolves (CR 608.2b); the -1/-1
// counter is put on the defending player's creature even if the first target
// had none to give.
//
// No simplification.
func init() {
	enters := effects.OldB242WhenThisEntersMinusCountersOnYourCreature("Decimator Beetle", 1)
	attacks := game.TriggeredAbility{
		Watches:     []game.EventKind{game.EventAttack},
		AppliesTo:   effects.ThisAttacked,
		TargetsFrom: effects.OldB242DecimatorTargets,
		Key:         "Decimator Beetle — remove a -1/-1 counter from target creature you control and put a -1/-1 counter on up to one target creature defending player controls",
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if mine, ok := ctx.ClauseTarget(0); ok && mine.Kind == game.TargetCard {
				if c, ok := g.LookupCardForEffect(mine.ID); ok && c.Counters[game.CounterMinusOne] > 0 {
					if err := (effects.AddCounter{Target: mine.ID, Kind: game.CounterMinusOne, N: -1}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if theirs, ok := ctx.ClauseTarget(1); ok && theirs.Kind == game.TargetCard {
				return effects.AddCounter{Target: theirs.ID, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
			}
			return nil
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "c56e6715-80bc-4f9d-86ef-7c31b1ef0a17",
		Name:         "Decimator Beetle",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters, attacks},
	})
}

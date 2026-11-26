package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exemplar of Strength — Creature — Human Warrior {1}{G}, 4/4 (Amonkhet, S82 OLD batch 242, #5535):
//
//	"When this creature enters, put three -1/-1 counters on target creature you
//	 control.
//	 Whenever this creature attacks, remove a -1/-1 counter from it. If you do,
//	 you gain 1 life."
//
// "If you do" is read as the ability resolves: with no -1/-1 counter left to
// remove, no life is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f732873e-0b34-4ccc-8b4c-9f02463ba57b",
		Name:         "Exemplar of Strength",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB242WhenThisEntersMinusCountersOnYourCreature("Exemplar of Strength", 3),
			effects.WheneverThisAttacks("Exemplar of Strength — remove a -1/-1 counter from it; if you do, you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					c, ok := g.LookupCardForEffect(ctx.Source())
					if !ok || c.Counters[game.CounterMinusOne] <= 0 || !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					if err := (effects.AddCounter{Target: ctx.Source(), Kind: game.CounterMinusOne, N: -1}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
				}),
		},
	})
}

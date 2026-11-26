package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chainer's Torment — Enchantment — Saga {3}{B} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — This Saga deals 2 damage to each opponent and you gain 2 life.
//	 III — Create an X/X black Nightmare Horror creature token, where X is
//	 half your life total, rounded up. It deals X damage to you."
//
// I and II are two chapters over one body: the Saga deals the damage to each
// opponent and its controller gains the 2 life. III reads your life total as
// the chapter resolves, makes the token at that size (a 0/0 template whose
// printed size is */*, set as it is created), and the token itself deals X
// damage to its controller. A life total of 0 or less makes X zero, and the
// 0/0 token dies (CR 704.5f).
//
// No simplification.
func init() {
	const name = "Chainer's Torment"
	drain := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		// One instance of damage to every opponent (CR 615.8).
		if err := g.DamageInstanceForEffect(func() error {
			for _, opp := range ctx.Opponents() {
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: opp, Amount: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
	}
	effects.Register(effects.Spec{
		OracleID:     "3c18113f-8d1e-4c63-a759-8cfc94e90876",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "this Saga deals 2 damage to each opponent and you gain 2 life"), drain),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "this Saga deals 2 damage to each opponent and you gain 2 life"), drain),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "create an X/X black Nightmare Horror creature token, where X is half your life total, rounded up; it deals X damage to you"),
				effects.OldB207NightmareHorror),
		},
	})
}

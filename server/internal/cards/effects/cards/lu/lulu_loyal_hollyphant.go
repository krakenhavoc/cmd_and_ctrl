package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lulu, Loyal Hollyphant — Legendary Creature — Elephant Angel {3}{W}, 3/2
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 At the beginning of your end step, if a permanent you controlled left
//	 the battlefield this turn, put a +1/+1 counter on each tapped creature
//	 you control, then untap them.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// Solemn Recruit's end-step revolt trigger (revolt.go; the intervening "if"
// is read as the end step begins and again as it resolves, CR 603.4). The
// tapped creatures are chosen as it resolves; each gets its counter, and
// once every counter has landed those same creatures untap (a prompt over
// one counter's replacements settles first). Choose a Background is deck
// construction (ADR 0144).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d99197b9-143a-462f-ace4-f35228ba082e",
		Name:            "Lulu, Loyal Hollyphant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfRevolt("Lulu, Loyal Hollyphant — put a +1/+1 counter on each tapped creature you control, then untap them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.Tapped {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					tail := ctx.ItemTail(luluUntapThem)
					tail.Carry = ids
					return effects.B13PutCounterOnEachThen(ctx, ids, tail)
				}),
		},
	})
}

// luluUntapThem is "then untap them": the creatures that were tapped as the
// ability resolved, whichever are still on the battlefield. An on-disk key.
var luluUntapThem = effects.CardTail("lulu-loyal-hollyphant/untap", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.Tail.Carry {
		if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

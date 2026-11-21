package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Of Herbs and Stewed Rabbit — Enchantment — Saga {2}{W} (Tales of
// Middle-earth Commander):
//
//	"I — Put a +1/+1 counter on up to one target creature. Create a Food
//	 token.
//	 II — Draw a card. Create a Food token.
//	 III — Create a 1/1 white Halfling creature token for each Food you
//	 control."
//
// Chapter I's target is optional and the Food is made whether or not it is
// still legal (CR 608.2b); chapter II's draw and Food are two plain steps;
// chapter III counts the Food permanents you control as the chapter
// resolves, tokens or not, by their layered subtype.
//
// No simplification.
func init() {
	const name = "Of Herbs and Stewed Rabbit"
	effects.Register(effects.Spec{
		OracleID:     "2efb908f-d069-4041-930b-5c1ccec00db2",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "put a +1/+1 counter on up to one target creature; create a Food token"),
				effects.UpToOneTargetCreature("up to one target creature"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.CreateToken{Template: effects.FoodToken(), N: 1}.Apply(ctx)
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "draw a card; create a Food token"),
				effects.Do(effects.DrawCards{N: 1}, effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "create a 1/1 white Halfling creature token for each Food you control"),
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.HasSubtype("Food") {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Halfling"), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

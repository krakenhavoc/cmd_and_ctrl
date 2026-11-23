package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of Eorl — Enchantment — Saga {3}{R}{W} (Tales of Middle-earth
// Commander):
//
//	"I — Create two 1/1 white Human Soldier creature tokens.
//	 II — Create two 2/2 red Human Knight creature tokens with trample and
//	 haste.
//	 III — Put an indestructible counter on up to one target Human. You
//	 become the monarch."
//
// Chapter III targets up to one Human (the target is optional but the
// monarch clause is not conditioned on it): the counter goes on it if it is
// still a legal target, and you become the monarch either way (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Oath of Eorl"
	effects.Register(effects.Spec{
		OracleID:     "3f8f72d8-5c3a-4f31-8741-3c1a55d9b739",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create two 1/1 white Human Soldier creature tokens"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 2})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create two 2/2 red Human Knight creature tokens with trample and haste"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Human Knight with trample and haste"), N: 2})),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "put an indestructible counter on up to one target Human; you become the monarch"),
				effects.TargetPermanent("up to one target Human", effects.Subtype("Human")).WithCount(0, 1),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterIndestructible, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.BecomeTheMonarch{}.Apply(ctx)
				}),
		},
	})
}

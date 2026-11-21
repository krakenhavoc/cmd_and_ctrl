package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oath of the Grey Host — Enchantment — Saga {3}{B} (The Lord of the Rings:
// Tales of Middle-earth, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — You and target opponent each create a Food token.
//	 II — Each opponent loses 3 life. Create a Treasure token.
//	 III — Create three tapped 1/1 white Spirit creature tokens with flying."
//
// Chapter I targets as it goes on the stack and you make your Food whatever
// happens to the target: with the opponent gone as it resolves, only the
// opponent's Food is lost (CR 608.2b). Chapter II drains every opponent, then
// makes the Treasure. Chapter III is three tapped Spirit tokens
// (a TokenSpec that enters tapped).
//
// No simplification.
func init() {
	const name = "Oath of the Grey Host"
	effects.Register(effects.Spec{
		OracleID:     "228fe7f0-54ad-411b-bedb-65a1bc2a099b",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "you and target opponent each create a Food token"),
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CreateToken{Template: effects.FoodToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.CreateToken{Controller: t.ID, Template: effects.FoodToken(), N: 1}.Apply(ctx)
						}
					}
					return nil
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "each opponent loses 3 life; create a Treasure token"),
				effects.Do(effects.EachOpponentLosesLife{N: 3}, effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "create three tapped 1/1 white Spirit creature tokens with flying"),
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("1/1 white Spirit with flying")).EntersTapped(), N: 3})),
		},
	})
}

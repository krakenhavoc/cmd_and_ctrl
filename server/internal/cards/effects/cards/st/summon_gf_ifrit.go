package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: G.F. Ifrit — Enchantment Creature — Saga Demon {2}{R}, 3/2
// (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II — You may discard a card. If you do, draw a card.
//	 III, IV — Add {R}."
//
// A Saga creature. Chapters I and II are Founding of Omashu's optional
// loot: the draw follows only a discard that happened. Chapters III and
// IV are Burn, Burn Tree and Fern's: one red mana into the pool as the
// chapter resolves, which empties at the end of the phase as usual.
//
// No simplification.
func init() {
	const name = "Summon: G.F. Ifrit"
	loot := func(n int) game.TriggeredAbility {
		label := effects.SagaChapterLabel(name, n, "you may discard a card; if you do, draw a card")
		return effects.ChapterTrigger(n, label, func(g *game.Game, item *game.StackItem) error {
			return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
		})
	}
	effects.Register(effects.Spec{
		OracleID:     "96d1bfc1-b80f-466f-b07d-df1ab361e8e0",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			loot(1),
			loot(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "add {R}"), effects.Do(effects.AddMana{Produced: "{R}"})),
			effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "add {R}"), effects.Do(effects.AddMana{Produced: "{R}"})),
		},
	})
}

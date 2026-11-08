package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
		label := SagaChapterLabel(name, n, "you may discard a card; if you do, draw a card")
		return ChapterTrigger(n, label, func(g *game.Game, item *game.StackItem) error {
			return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
		})
	}
	Register(Spec{
		OracleID:     "96d1bfc1-b80f-466f-b07d-df1ab361e8e0",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			loot(1),
			loot(2),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "add {R}"), Do(AddMana{Produced: "{R}"})),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "add {R}"), Do(AddMana{Produced: "{R}"})),
		},
	})
}

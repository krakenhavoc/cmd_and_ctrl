package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault 75: Middle School — Enchantment — Saga {2}{W}{W} (Fallout,
// #5469):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Exile all creatures with power 4 or greater.
//	 II, III — Put a +1/+1 counter on each creature you control."
//
// Chapter I is Farewell's mass exile (ExileAllMatching) over every
// creature whose power is 4 or greater as the chapter resolves, any
// player's, on the layered board. Chapters II and III are Abandoned Air
// Temple's counter on each creature you control
// (b11PutCounterOnEachCreatureYouControl). The lore counters and the
// sacrifice after III are the engine's Saga lifecycle (CR 714).
//
// No simplification.
func init() {
	const name = "Vault 75: Middle School"
	counters := func(n int) game.TriggeredAbility {
		return ChapterTrigger(n, SagaChapterLabel(name, n, "put a +1/+1 counter on each creature you control"),
			b11PutCounterOnEachCreatureYouControl)
	}
	Register(Spec{
		OracleID:     "e81fa87f-80ce-4db6-940e-d30309bc70e6",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(ChapterTrigger(1, SagaChapterLabel(name, 1, "exile all creatures with power 4 or greater"),
				func(g *game.Game, item *game.StackItem) error {
					return ExileAllMatching{Match: And(Creature(), PowerGE(4))}.Apply(NewContext(g, item))
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}}),
			counters(2),
			counters(3),
		},
	})
}

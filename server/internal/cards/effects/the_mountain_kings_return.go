package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Mountain-king's Return — Enchantment — Saga {2}{W}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Recruit. (Draw a card, then discard a card. If you discarded a
//	 nonland card, create a 1/1 white Human Soldier creature token.)
//	 II — Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield.
//	 III — Put a +1/+1 counter on up to one target creature."
//
// The lore counters and the sacrifice are the engine's (CR 714); the
// chapters are ordinary triggers. Chapter II with no such card in your
// graveyard is removed for want of a target (CR 603.3d); chapter III may
// choose no creature.
//
// No simplification. The Hobbit (#4475).
const theMountainKingsReturnName = "The Mountain-king's Return"

func init() {
	Register(Spec{
		OracleID:     "32ad5b3e-92c0-45be-b2e4-6f1794552f36",
		Name:         theMountainKingsReturnName,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(theMountainKingsReturnName, 1, "recruit"), Do(Recruit{})),
			ChapterTriggerTargeting(2, SagaChapterLabel(theMountainKingsReturnName, 2, "return a creature card with mana value 3 or less"),
				TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", YouOwn(), Creature(), ManaValueLE(3)),
				sheoldredReanimate),
			ChapterTriggerTargeting(3, SagaChapterLabel(theMountainKingsReturnName, 3, "a +1/+1 counter on up to one creature"),
				TargetCreature("up to one target creature").WithCount(0, 1),
				PlusOneCounterOnTheTarget),
		},
	})
}

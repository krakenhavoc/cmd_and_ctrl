package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amok Time — Enchantment — Saga {2}{G} (Star Trek, #4973):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Target creature you control fights up to one target creature an
//	     opponent controls.
//	 II — You gain 3 life.
//	 III — Return target creature card from your graveyard to your hand."
//
// Chapter I has two target clauses, chosen as it goes on the stack: the
// creature you control (needed, so with none the chapter is removed, CR
// 603.3d) and up to one opponent's creature. The fight is
// slotZeroFightsSlotOne: with the second clause left empty, or either
// creature no longer legal, nothing fights (CR 701.14b). Chapter II is a
// plain life gain; chapter III is Kraven's Last Hunt's return to hand.
//
// No simplification.
func init() {
	const name = "Amok Time"
	Register(Spec{
		OracleID:     "98a752ff-f041-4032-9f3f-c6c4007f10f1",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "target creature you control fights up to one target creature an opponent controls"),
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
				),
				slotZeroFightsSlotOne),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "you gain 3 life"), Do(GainLife{Amount: 3})),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "return target creature card from your graveyard to your hand"),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
				returnTargetCardToHand),
		},
	})
}

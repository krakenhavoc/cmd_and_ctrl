package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "98a752ff-f041-4032-9f3f-c6c4007f10f1",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "target creature you control fights up to one target creature an opponent controls"),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1),
				),
				effects.SlotZeroFightsSlotOne),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "you gain 3 life"), effects.Do(effects.GainLife{Amount: 3})),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "return target creature card from your graveyard to your hand"),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTargetCardToHand),
		},
	})
}

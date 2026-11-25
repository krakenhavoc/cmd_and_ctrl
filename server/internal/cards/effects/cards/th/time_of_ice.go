package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time of Ice — Enchantment — Saga {3}{U} (Dominaria):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Tap target creature an opponent controls. It doesn't untap during
//	 its controller's untap step for as long as you control this Saga.
//	 III — Return all tapped creatures to their owners' hands."
//
// I and II are Dungeon Geists' tap and hold with the Saga as the thing you
// must keep controlling: the target is chosen as the chapter goes on the
// stack, and losing control of the Saga (or the Saga leaving after III) ends
// the hold. III bounces every tapped creature, whoever controls it, as the
// chapter resolves; the Saga is sacrificed only after it has resolved
// (CR 714.4).
//
// No simplification.
func init() {
	const name = "Time of Ice"
	hold := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		return effects.TapAndHoldWhileYouControlThis(ctx, effects.HoldTargetIDs(ctx))
	}
	opp := func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}
	effects.Register(effects.Spec{
		OracleID:     "93e0fd89-8a0d-4d97-aa22-aafe83d53515",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "tap target creature an opponent controls; it doesn't untap while you control this Saga"), opp(), hold),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "tap target creature an opponent controls; it doesn't untap while you control this Saga"), opp(), hold),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "return all tapped creatures to their owners' hands"),
				effects.Do(effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.B751Tapped())})),
		},
	})
}

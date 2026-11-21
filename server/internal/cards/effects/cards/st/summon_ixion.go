package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Ixion — Enchantment Creature — Saga Unicorn {2}{W}, 3/3:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Aerospark — Exile target creature an opponent controls until
//	 this Saga leaves the battlefield.
//	 II, III — Put a +1/+1 counter on each of up to two target creatures
//	 you control. You gain 2 life.
//	 First strike"
//
// A Saga creature (ChapterTriggerTargeting). Chapter I is the Oblivion
// Ring body (ExileUntil.ThisLeaves): the creature comes back when the
// Saga leaves, which for an unanswered Ixion is when it is sacrificed
// after chapter III (CR 714.4). Chapters II and III gain the life
// whether or not a creature was picked.
//
// No simplification.
func init() {
	// The life gain waits for a prompt the counters ask (CR 608.2c, #4155).
	counters := effects.Do(
		effects.Instruction(func(ctx *effects.Context) error { return effects.SupportEffect(ctx.Game, ctx.Item) }),
		effects.GainLife{Amount: 2},
	)
	upToTwo := func() *game.TargetSpec {
		return effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2)
	}
	const name = "Summon: Ixion"
	effects.Register(effects.Spec{
		OracleID:        "1ccb8122-9135-4b8d-b1d1-18997fd7e9d2",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "Aerospark — exile target creature an opponent controls until this Saga leaves the battlefield"),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				effects.ExileChosenTargetUntilThisLeaves("Summon: Ixion — the exiled creature returns when the Saga leaves the battlefield")),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "a +1/+1 counter on each of up to two creatures you control; gain 2 life"), upToTwo(), counters),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "a +1/+1 counter on each of up to two creatures you control; gain 2 life"), upToTwo(), counters),
		},
	})
}

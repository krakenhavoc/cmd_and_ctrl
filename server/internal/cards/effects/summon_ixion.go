package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	counters := func(g *game.Game, item *game.StackItem) error {
		if err := SupportEffect(g, item); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
	}
	upToTwo := func() *game.TargetSpec {
		return TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2)
	}
	const name = "Summon: Ixion"
	Register(Spec{
		OracleID:        "1ccb8122-9135-4b8d-b1d1-18997fd7e9d2",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "Aerospark — exile target creature an opponent controls until this Saga leaves the battlefield"),
				TargetCreature("target creature an opponent controls", OpponentControls()),
				exileChosenTargetUntilThisLeaves("Summon: Ixion — the exiled creature returns when the Saga leaves the battlefield")),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "a +1/+1 counter on each of up to two creatures you control; gain 2 life"), upToTwo(), counters),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "a +1/+1 counter on each of up to two creatures you control; gain 2 life"), upToTwo(), counters),
		},
	})
}

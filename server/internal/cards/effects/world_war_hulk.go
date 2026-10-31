package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// World War Hulk — {3}{G}{G} Enchantment — Saga (Marvel Super Heroes, #4455,
// #4615):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — The next red or green creature spell you cast this turn can be cast
//	 without paying its mana cost.
//	 II — Put three +1/+1 counters on target creature you control.
//	 III — Choose target creature you control. Until end of turn, double its
//	 power and toughness and it gains trample."
//
// I is Apex Observatory's free-cast promise (GrantNextSpellPromise with
// WithoutPayingManaCost, CR 118.9) with a colour clause on the filter:
// PermissionFilter.ColorsAny "RG" with CreatureOnly, so a red or green
// creature spell qualifies (a multicoloured one needs one of the two, CR
// 105.2b), and a blue creature spell or a red noncreature spell does not and
// leaves the promise for the next one. It lasts until end of turn and is
// spent by the first creature spell it matches, whichever way that spell was
// paid for.
//
// II is three counters on the chosen creature. III reads the creature's power
// and toughness as it resolves (CR 701.10b: doubling adds the current value)
// and gives it trample until end of turn; a creature gone from the
// battlefield by then is neither doubled nor given anything (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c4da760-8913-46ca-b00c-bba6d877d482",
		Name:         "World War Hulk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("World War Hulk", 1, "the next red or green creature spell you cast this turn can be cast without paying its mana cost"),
				Do(GrantNextSpellPromise{From: "World War Hulk", Promise: game.NextSpellPromise{
					Filter:                game.PermissionFilter{CreatureOnly: true, ColorsAny: "RG"},
					WithoutPayingManaCost: true,
					Text:                  "The next red or green creature spell you cast this turn can be cast without paying its mana cost.",
				}})),
			ChapterTriggerTargeting(2, SagaChapterLabel("World War Hulk", 2, "three +1/+1 counters on target creature you control"),
				TargetCreature("target creature you control", YouControl()), worldWarHulkCounters),
			ChapterTriggerTargeting(3, SagaChapterLabel("World War Hulk", 3, "double target creature's power and toughness and it gains trample until end of turn"),
				TargetCreature("target creature you control", YouControl()), worldWarHulkDouble),
		},
	})
}

// worldWarHulkCounters is chapter II.
func worldWarHulkCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
		}
	}
	return nil
}

// worldWarHulkDouble is chapter III: the doubling is read now, then trample.
func worldWarHulkDouble(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := DoublePowerAndToughnessOfEachUntilEOT(ctx, []uuid.UUID{t.ID}, "World War Hulk — double power and toughness"); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "World War Hulk — gains trample"}.Apply(ctx)
	}
	return nil
}

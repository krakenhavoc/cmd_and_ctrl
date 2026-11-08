package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revival of the Ancestors — Enchantment — Saga {1}{W}{B}{G} (Tarkir:
// Dragonstorm, #4463):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create three 1/1 white Spirit creature tokens.
//	 II — Distribute three +1/+1 counters among one, two, or three target
//	      creatures you control.
//	 III — Creatures you control gain trample and lifelink until end of
//	       turn."
//
// Chapter II is Armament Dragon's divided counters as a chapter ability:
// the split is announced with the targets (CR 601.2d) and a target gone by
// resolution gets none of its share. Chapter III's creatures are fixed as
// it resolves (CR 611.2c), so a creature that enters later that turn gains
// nothing.
//
// No simplification.
func init() {
	const name = "Revival of the Ancestors"
	Register(Spec{
		OracleID:     "5a848c1f-d11a-42fb-b9ce-663901fc9555",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create three 1/1 white Spirit creature tokens"),
				Do(CreateToken{Template: revivalOfTheAncestorsSpirit(), N: 3})),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "distribute three +1/+1 counters"),
				TargetCreature("one, two, or three target creatures you control", YouControl()).
					WithCount(1, 3).Dividing(Divide(3)),
				jadeSeedstonesCounters),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "creatures you control gain trample and lifelink until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"trample", "lifelink"},
						Label:    "Revival of the Ancestors — trample and lifelink until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// revivalOfTheAncestorsSpirit is the plain 1/1 white Spirit, a token the
// table has no row for.
func revivalOfTheAncestorsSpirit() game.Card {
	return game.Card{
		Name:      "Spirit",
		TypeLine:  "Token Creature — Spirit",
		Colors:    []string{"W"},
		Power:     1,
		Toughness: 1,
	}
}

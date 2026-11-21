package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5a848c1f-d11a-42fb-b9ce-663901fc9555",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create three 1/1 white Spirit creature tokens"),
				effects.Do(effects.CreateToken{Template: revivalOfTheAncestorsSpirit(), N: 3})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "distribute three +1/+1 counters"),
				effects.TargetCreature("one, two, or three target creatures you control", effects.YouControl()).
					WithCount(1, 3).Dividing(effects.Divide(3)),
				effects.JadeSeedstonesCounters),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "creatures you control gain trample and lifelink until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"trample", "lifelink"},
						Label:    "Revival of the Ancestors — trample and lifelink until end of turn",
					}.Apply(effects.NewContext(g, item))
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

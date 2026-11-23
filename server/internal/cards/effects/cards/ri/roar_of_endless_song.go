package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roar of Endless Song — Enchantment — Saga {2}{G}{U}{R} (Tarkir:
// Dragonstorm, #4463):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Create a 5/5 green Elephant creature token.
//	 III — Double the power and toughness of each creature you control
//	       until end of turn."
//
// Chapters I and II are two entries with the same effect, as printed.
// Chapter III is Unnatural Growth's doubling: each creature you control as
// it resolves (CR 611.2c) gets +X/+X where X is its power and toughness
// then (CR 701.10b), so a creature that enters later that turn is not
// doubled.
//
// No simplification.
func init() {
	const name = "Roar of Endless Song"
	elephant := effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 green Elephant"), N: 1})
	effects.Register(effects.Spec{
		OracleID:     "a3570093-3c16-42e7-a83d-2ff611a891b5",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 5/5 green Elephant creature token"), elephant),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a 5/5 green Elephant creature token"), elephant),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "double the power and toughness of each creature you control"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DoublePowerAndToughnessOfEachUntilEOT(effects.NewContext(g, item), effects.TdmCreaturesYouControl(g, item.Controller),
						"Roar of Endless Song — double power and toughness")
				}),
		},
	})
}

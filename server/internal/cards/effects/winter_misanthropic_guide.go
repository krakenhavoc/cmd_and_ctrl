package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winter, Misanthropic Guide — Legendary Creature — Human Warlock
// {1}{B}{R}{G}, 3/4:
//
//	"Ward {2}
//	 At the beginning of your upkeep, each player draws two cards.
//	 Delirium — As long as there are four or more card types among
//	 cards in your graveyard, each opponent's maximum hand size is
//	 equal to seven minus the number of those card types."
//
// The delirium clause is a maximum-hand-size static that applies only
// while its condition holds and reads its number when asked
// (HandSizeStatic.Dynamic, ADR 0113 §3, #2074). It keeps Winter's own
// timestamp in CR 613.11's order whenever delirium turns on (the
// 2024-09-20 ruling: a Reliquary Tower that entered after Winter always
// wins; one that entered before it loses while you have delirium). Seven
// minus eight or nine types is below zero, which reads as zero
// (CR 107.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a9baafb-bffd-4e28-bbe1-b4154cd86bf3",
		Name:         "Winter, Misanthropic Guide",
		Completeness: CompletenessFull,
		HandSize: []game.HandSizeStatic{{
			Players: game.HandSizeEachOpponent,
			Kind:    game.HandSizeSet,
			Dynamic: func(g *game.Game, source *game.Card) (int, bool) {
				types := b16CardTypesInGraveyard(g, source.Controller)
				if types < 4 {
					return 0, false
				}
				return 7 - types, true
			},
		}},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Winter, Misanthropic Guide — ward {2}"),
			AtYourUpkeep("Winter, Misanthropic Guide — each player draws two cards", func(g *game.Game, item *game.StackItem) error {
				return b05EachPlayerDraws(g, item, 2)
			}),
		},
	})
}

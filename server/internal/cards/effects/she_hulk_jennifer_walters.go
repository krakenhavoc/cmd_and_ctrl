package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// She-Hulk, Jennifer Walters — Legendary Creature — Gamma Berserker
// Hero {3}{R}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 {2}{R}, Sacrifice a land: Draw a card and put a +1/+1 counter on
//	 She-Hulk."
//
// The land is a sacrifice cost (Hammer of Purphoros's reader), so it is
// gone before a response can stop the ability. The counter goes on
// She-Hulk only if she is still on the battlefield when it resolves
// (CounterOnThis); the draw happens regardless.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d62b8609-78a5-43ef-8e6e-dbfe8c7453ca",
		Name:            "She-Hulk, Jennifer Walters",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice a land: Draw a card and put a +1/+1 counter on She-Hulk.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{R}"), game.AbilityCost{SacrificeOther: b29SacrificeALand()}),
			Effect:  Do(DrawCards{N: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}

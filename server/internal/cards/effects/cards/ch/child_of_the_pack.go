package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Child of the Pack // Savage Packmate — {2}{R}{G} Creature — Human
// Werewolf 2/5 // Creature — Werewolf 5/5 (#2586, ADR 0132):
//
//	Front: "{2}{R}{G}: Create a 2/2 green Wolf creature token.
//	        Daybound"
//	Back:  "Trample
//	        Other creatures you control get +1/+0.
//	        Nightbound"
//
// The back face's anthem is a tribe filter with no tribe: every other
// creature its controller controls.
//
// No simplification.
func init() {
	const oracle = "a9c5b155-7c09-4100-9679-8fca2b5e9222"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Child of the Pack",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}{G}: Create a 2/2 green Wolf creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{2}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Savage Packmate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "nightbound"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 0),
		},
	})
}

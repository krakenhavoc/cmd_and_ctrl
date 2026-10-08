package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        oracle,
		Name:            "Child of the Pack",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Activated: []ActivatedAbility{{
			Label: "{2}{R}{G}: Create a 2/2 green Wolf creature token.",
			Cost:  ManaCost("{2}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Savage Packmate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "nightbound"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 0),
		},
	})
}

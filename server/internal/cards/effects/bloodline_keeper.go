package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodline Keeper // Lord of Lineage — {2}{B}{B} Creature — Vampire
// 3/3 // Creature — Vampire 5/5 (ADR 0132 transform):
//
//	Front: "Flying
//	        {T}: Create a 2/2 black Vampire creature token with flying.
//	        {B}: Transform this creature. Activate only if you control
//	        five or more Vampires."
//	Back:  "Flying
//	        Other Vampire creatures you control get +2/+2.
//	        {T}: Create a 2/2 black Vampire creature token with flying."
//
// "Five or more Vampires" counts permanents with the type, the Keeper
// included, and is an activation instruction (CR 602.1b): checked when
// the ability is activated, never at resolution, so the Keeper still
// turns over if a Vampire dies in response. It is read through the
// source's text (CR 612.2), so a changeling counts. The transform is
// TransformThis, which ignores an instruction aimed at a Keeper that has
// already turned over since the ability went on the stack (CR 701.27f).
//
// The token is a black 2/2 flier, a Vampire, so it counts toward the five
// and takes the Lord's +2/+2 as soon as it lands. The Lord's anthem is
// "other Vampire creatures you control" (Others, YoursOnly): the lord
// itself is untouched and an opponent's Vampires are not pumped.
//
// No simplification.
func init() {
	const oracle = "434e82b5-abd4-41ac-a5b1-2dd01c667374"
	makeVampire := func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 black Vampire with flying"), N: 1}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Bloodline Keeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Create a 2/2 black Vampire creature token with flying.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
				Effect:  makeVampire,
			},
			{
				Label:     "{B}: Transform this creature. Activate only if you control five or more Vampires.",
				Cost:      ManaCost("{B}"),
				Purpose:   game.Purpose{Answers: game.AnswerPump},
				Condition: ControlsAtLeast(5, func(c game.Card) bool { return c.HasSubtypeWord("Vampire") }),
				Effect:    Do(TransformThis{}),
			},
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Lord of Lineage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}, 2, 2),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Create a 2/2 black Vampire creature token with flying.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Effect:  makeVampire,
		}},
	})
}

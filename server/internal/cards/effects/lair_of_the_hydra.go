package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lair of the Hydra — Land (Adventures in the Forgotten Realms):
//
//	"If you control two or more other lands, this land enters tapped.
//	 {T}: Add {G}.
//	 {X}{G}: Until end of turn, this land becomes an X/X green Hydra
//	 creature. It's still a land. X can't be 0."
//
// The animation is afrB17Manland with the X announced on activation
// (CR 107.3, item.XValue) as both base power and toughness; "X can't be 0"
// is MinX(1), so a zero announcement is refused outright.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "126e9140-2c05-4c00-8b01-5653456c736a",
		Name:          "Lair of the Hydra",
		Completeness:  CompletenessFull,
		XMatters:      true,
		Replacements:  []game.ReplacementEffect{afrB17EntersTappedWithTwoOtherLands()},
		ManaAbilities: afrB17TapForMana("G"),
		Activated: []ActivatedAbility{{
			Label:   "{X}{G}: Until end of turn, this land becomes an X/X green Hydra creature. It's still a land. X can't be 0.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    Plus(ManaCost("{X}{G}"), MinX(1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := max(item.XValue, 0)
				return afrB17Manland(g, item, "Lair of the Hydra", "Hydra", "G", x, x, nil, "")
			},
		}},
	})
}

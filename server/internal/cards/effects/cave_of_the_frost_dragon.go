package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cave of the Frost Dragon — Land (Adventures in the Forgotten Realms):
//
//	"If you control two or more other lands, this land enters tapped.
//	 {T}: Add {W}.
//	 {4}{W}: This land becomes a 3/4 white Dragon creature with flying until
//	 end of turn. It's still a land."
//
// The animation is afrB17Manland, flying as a layer-6 keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "1e4146d2-cfa0-4f5e-9761-3c83519b90c3",
		Name:          "Cave of the Frost Dragon",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{afrB17EntersTappedWithTwoOtherLands()},
		ManaAbilities: afrB17TapForMana("W"),
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: This land becomes a 3/4 white Dragon creature with flying until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{4}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return afrB17Manland(g, item, "Cave of the Frost Dragon", "Dragon", "W", 3, 4, []string{"flying"}, "")
			},
		}},
	})
}

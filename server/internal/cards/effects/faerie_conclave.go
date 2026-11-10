package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Conclave — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {1}{U}: This land becomes a 2/1 blue Faerie creature with flying
//	 until end of turn. It's still a land."
//
// A manland built like Creeping Tar Pit, through cmdr2B13Animate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c25f6b1-8fb3-4406-9605-0282d2dbbcec",
		Name:         "Faerie Conclave",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: This land becomes a 2/1 blue Faerie creature with flying until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return cmdr2B13Animate(NewContext(g, item), item.SourceCardID,
					"Faerie Conclave — a 2/1 blue Faerie creature with flying until end of turn",
					2, 1, "Faerie", []string{"U"}, "flying")
			},
		}},
	})
}

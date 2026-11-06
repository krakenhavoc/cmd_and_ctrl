package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ahriman — Creature — Eye Horror {2}{B}, 2/2 (Final Fantasy, #4462):
//
//	"Flying, deathtouch
//	 {3}, Sacrifice another creature or artifact: Draw a card."
//
// The sacrifice is part of the cost (CR 602.2), so the creature or
// artifact is gone before the ability can be answered, and "another"
// keeps Ahriman from paying with itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "37d1d507-25dc-4cd0-b51e-2f1e9b24832c",
		Name:            "Ahriman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice another creature or artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{3}"), game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			}),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}

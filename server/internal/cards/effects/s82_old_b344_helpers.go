package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b344_helpers.go — S82 OLD batch 344 (#5554).

// KhansBannerSpec is the shape the five Khans of Tarkir Banners share:
//
//	"{T}: Add {A}, {B}, or {C}.
//	 {A}{B}{C}, {T}, Sacrifice this artifact: Draw a card."
//
// `colors` are the three colour letters in the card's printed order. The
// mana ability is the Cluestone shape (a mana ability that adds one of
// three colours); the second ability taps and sacrifices as costs, so
// removal in response cannot stop the draw.
func KhansBannerSpec(oracleID, name string, colors [3]string) Spec {
	cost := "{" + colors[0] + "}{" + colors[1] + "}{" + colors[2] + "}"
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + colors[0] + "|" + colors[1] + "|" + colors[2] + "}",
			Label:    "Add {" + colors[0] + "}, {" + colors[1] + "}, or {" + colors[2] + "}",
		}},
		Activated: []ActivatedAbility{{
			Label:   cost + ", {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: cost},
			Effect:  Do(DrawCards{N: 1}),
		}},
	}
}

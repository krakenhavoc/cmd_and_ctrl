package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Redrock Sentinel — Artifact Creature — Golem {3}, 2/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"Defender
//	 {2}, {T}, Sacrifice a land: Draw a card and create a Treasure token.
//	 (It's an artifact with "{T}, Sacrifice this token: Add one mana of
//	 any color.")"
//
// Zuran Orb's "Sacrifice a land" cost (b08SacrificeALand, the same
// picker every sac-outlet cost opens) beside the mana and the tap, all
// paid at announce (CR 602.2b). The effect is Reckless Lackey's: a card,
// then a Treasure. A Sentinel with no land to sacrifice can't activate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51a6819d-c8fb-4f52-ab0c-4ae845728ed2",
		Name:            "Redrock Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a land: Draw a card and create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost(), b08SacrificeALand()),
			Effect:  otjRDrawThenTreasure(),
		}},
	})
}

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
// The land sacrifice is Seismic Monstrosaur's cost; {T} is a creature's
// tap cost, so the Sentinel can't activate the turn it arrives. The body
// is Reckless Lackey's (drawACardAndATreasure).
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
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Tokens: 1},
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeN(1, "a land", Land())),
			Effect:  drawACardAndATreasure(),
		}},
	})
}

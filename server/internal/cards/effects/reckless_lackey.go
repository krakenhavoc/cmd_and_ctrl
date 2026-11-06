package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Lackey — Creature — Goblin Pirate {R}, 1/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"First strike, haste
//	 {2}{R}, Sacrifice this creature: Draw a card and create a Treasure
//	 token. (It's an artifact with "{T}, Sacrifice this token: Add one
//	 mana of any color.")"
//
// The activation is Silversmote Ghoul's self-sacrifice row with a Treasure
// after the draw (drawACardAndATreasure, shared with Redrock Sentinel).
// The sacrifice is a cost, so the card is drawn even if the ability is
// the last thing the Lackey does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73733fa5-d292-40ec-8f17-05ba5feb40ab",
		Name:            "Reckless Lackey",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this creature: Draw a card and create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Tokens: 1},
			Cost:    Plus(ManaCost("{2}{R}"), SacrificeThis()),
			Effect:  drawACardAndATreasure(),
		}},
	})
}

// drawACardAndATreasure is "Draw a card and create a Treasure token." as
// an activated ability's whole effect, under its controller (Reckless
// Lackey, Redrock Sentinel).
func drawACardAndATreasure() Effect {
	return Do(DrawCards{N: 1}, CreateToken{Template: TreasureToken(), N: 1})
}

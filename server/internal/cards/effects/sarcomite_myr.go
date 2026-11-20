package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarcomite Myr — Artifact Creature — Phyrexian Myr {2}{U}, 2/1
// (Future Sight, #5594):
//
//	"{2}: This creature gains flying until end of turn.
//	 {2}, Sacrifice this creature: Draw a card."
//
// Two activated abilities from shared bodies: the self keyword grant until
// end of turn and Heart Warden's sacrifice-to-draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e78fd8f6-eca9-4107-aee7-a65689cee2f4",
		Name:         "Sarcomite Myr",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{2}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Sarcomite Myr — flying until end of turn"),
			},
			{
				Label:   "{2}, Sacrifice this creature: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
				Effect:  ebDrawACard,
			},
		},
	})
}

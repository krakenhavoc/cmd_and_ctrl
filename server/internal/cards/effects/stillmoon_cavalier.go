package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stillmoon Cavalier — Creature — Zombie Knight {1}{W/B}{W/B}, 2/1 (Eventide, #5590):
//
//	"Protection from white and from black
//	 {W/B}: This creature gains flying until end of turn.
//	 {W/B}: This creature gains first strike until end of turn.
//	 {W/B}{W/B}: This creature gets +1/+0 until end of turn."
//
// Animar's two protections and Order of Leitbur's rows, each paid with
// white or black mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c430b5a-83b6-4f4f-b1d3-5737e1f61947",
		Name:            "Stillmoon Cavalier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from white", "protection from black"},
		Activated: []ActivatedAbility{
			{
				Label:   "{W/B}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{W/B}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Stillmoon Cavalier — flying until end of turn"),
			},
			{
				Label:   "{W/B}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{W/B}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "Stillmoon Cavalier — first strike until end of turn"),
			},
			{
				Label:   "{W/B}{W/B}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{W/B}{W/B}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Stillmoon Cavalier — +1/+0 until end of turn"),
			},
		},
	})
}

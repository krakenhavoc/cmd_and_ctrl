package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Truefire Paladin — Creature — Human Knight {R}{W}, 2/2 (Gatecrash, #5565):
//
//	"Vigilance
//	 {R}{W}: This creature gets +2/+0 until end of turn.
//	 {R}{W}: This creature gains first strike until end of turn."
//
// Vigilance is the engine keyword; the two abilities are Freejam
// Regent's pump and Burning Shield Askari's grant, each pinned to this
// creature until cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2a8a239c-7cc7-48d5-b34b-60b0da4e4b20",
		Name:            "Truefire Paladin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}{W}: This creature gets +2/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}{W}"),
				Effect:  thisGetsUntilEndOfTurn(2, 0, "Truefire Paladin — +2/+0 until end of turn"),
			},
			{
				Label:   "{R}{W}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
				Cost:    ManaCost("{R}{W}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("first strike", "Truefire Paladin — first strike until end of turn"),
			},
		},
	})
}

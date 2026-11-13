package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellkite Overlord — Creature — Dragon {4}{B}{R}{R}{G}, 8/8 (From the
// Vault: Dragons, #5589):
//
//	"Flying, trample, haste
//	 {R}: This creature gets +1/+0 until end of turn.
//	 {B}{G}: Regenerate this creature."
//
// Shivan Dragon's {R} pump and Albino Troll's regeneration (Regenerate
// on the ability's own source, CR 701.19) on one body. Neither ability
// taps, so each can be activated any number of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d05e441-84d4-482f-95a8-7842e776cad2",
		Name:            "Hellkite Overlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "haste"},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}"),
				Effect:  thisGetsUntilEndOfTurn(1, 0, "Hellkite Overlord — +1/+0"),
			},
			{
				Label:   "{B}{G}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{B}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

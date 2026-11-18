package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jenara, Asura of War — Legendary Creature — Angel {G}{W}{U}, 3/3:
//
//	"Flying
//	 {1}{W}: Put a +1/+1 counter on Jenara."
//
// The counter goes on Jenara as the ability resolves, and nothing happens
// if she has left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d06cd670-7ffc-4295-97d1-0eff042fb6d5",
		Name:            "Jenara, Asura of War",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Put a +1/+1 counter on Jenara.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{W}"),
			Effect:  putCounterOnSelf,
		}},
	})
}

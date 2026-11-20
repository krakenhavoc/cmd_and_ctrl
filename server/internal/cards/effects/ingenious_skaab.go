package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingenious Skaab — Creature — Zombie Horror {2}{U}, 2/3 (Eldritch Moon,
// #5541):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 {U}: This creature gets +1/-1 until end of turn."
//
// Prowess is an engine keyword carried as printed data; the activation is
// Frostburn Weird's at {U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9107ece7-a263-4203-ab09-069193cc4902",
		Name:            "Ingenious Skaab",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Activated: []ActivatedAbility{{
			Label:   "{U}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{U}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Ingenious Skaab — +1/-1"),
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Stalwart — Creature — Elephant Soldier {3}{W}{W}, 3/3 (Fifth Dawn, #5604):
//
//	"Vigilance
//	 {W}: This creature gets +0/+1 until end of turn."
//
// Honor Guard's {W}: +0/+1 on a vigilant body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6b2eb774-08f3-46cb-aeca-2ab82fdd5f03",
		Name:            "Loxodon Stalwart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{W}: This creature gets +0/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Toughness: 1}},
			Cost:    ManaCost("{W}"),
			Effect:  thisGetsUntilEndOfTurn(0, 1, "Loxodon Stalwart — +0/+1 until end of turn"),
		}},
	})
}

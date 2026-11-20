package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Char-Rumbler — Creature — Elemental {2}{R}{R}, -1/3 (Future Sight, #5594):
//
//	"Double strike
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Furnace Spirit's firebreathing on a double striker. The printed power of
// -1 rides game.Card from Scryfall; each {R} adds one until end of turn,
// so the first activation brings it to 0 and the second makes it deal
// damage in both combat damage steps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "758aeca5-6ef3-4463-a65a-66b39289cc86",
		Name:            "Char-Rumbler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Char-Rumbler — +1/+0 until end of turn"),
		}},
	})
}

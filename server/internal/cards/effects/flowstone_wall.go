package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flowstone Wall — Creature — Wall {2}{R}, 0/6 (Nemesis, #5618):
//
//	"Defender (This creature can't attack.)
//	 {R}: This creature gets +1/-1 until end of turn."
//
// Defender is a printed keyword, enforced at the attack declaration
// (CR 702.3b). The pump is Flowstone Mauler's: the shared
// thisGetsUntilEndOfTurn body at +1/-1 for {R}, ending in the cleanup
// step (CR 514.2), so the Wall can trade toughness for power as a
// blocker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5fa86b45-1d5b-42df-8811-6a10903f3039",
		Name:            "Flowstone Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Flowstone Wall — +1/-1 until end of turn"),
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Colos Yearling — Creature — Goat Beast {2}{R}, 1/1 (Urza's Destiny,
// #5621):
//
//	"Mountainwalk
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Vug Lizard's mountainwalk, which the block check reads off the
// defending player's lands after layer 4, plus a firebreathing row
// through the shared thisGetsUntilEndOfTurn body: no {T}, so it can be
// pumped any number of times, and the boosts end in the cleanup step
// (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "50cf702b-5647-415e-b9a8-56fa0e50da6d",
		Name:            "Colos Yearling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Colos Yearling — +1/+0 until end of turn"),
		}},
	})
}

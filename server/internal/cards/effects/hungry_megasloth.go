package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hungry Megasloth — Creature — Sloth Beast {2}{G}, 3/3 (Foundations
// Jumpstart, #5464):
//
//	"Reach (This creature can block creatures with flying.)
//	 {2}, {T}: Put a +1/+1 counter on this creature."
//
// Tender Wildguide's counter ability with {2} added to the tap: a
// summoning-sick Megasloth can't use it (CR 302.6), and a Megasloth that
// left the battlefield in response gets nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "90aacd34-ff76-4e7e-bdb3-9427d431b6e5",
		Name:            "Hungry Megasloth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect:  putCounterOnSelf,
		}},
	})
}

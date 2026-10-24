package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blorbian Buddy — Creature — Alien Guest {G}, 1/1:
//
//	"Trample
//	 {G}, {T}: You get {TK} (a ticket counter)."
//
// The ticket is a player counter (CR 107.17, ADR 0208). {T} in the cost
// means summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ebd06311-4aa6-4db9-a90a-666cc4e744d1",
		Name:            "Blorbian Buddy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: You get {TK}.",
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(GetTickets{N: 1}),
		}},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "ebd06311-4aa6-4db9-a90a-666cc4e744d1",
		Name:            "Blorbian Buddy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: You get {TK}.",
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.GetTickets{N: 1}),
		}},
	})
}

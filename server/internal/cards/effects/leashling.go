package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leashling — Artifact Creature — Dog {6}, 3/3:
//
//	"Put a card from your hand on top of your library: Return this
//	 creature to its owner's hand."
//
// ADR 0109 §7 (#1902): "Put a card from your hand on top of your
// library" is a cost the activator names at announce (CR 602.2b), never
// the card being activated. It is not a discard, and the activator still
// knows the card they put there.
// The return reads the Leashling as the object the ability came from
// (CR 400.7): one that left and came back in response is a new object
// and stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8460abd9-ab5d-4a7e-8132-185b4a310190",
		Name:         "Leashling",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Put a card from your hand on top of your library: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    PutACardFromHandOnTop(),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
}

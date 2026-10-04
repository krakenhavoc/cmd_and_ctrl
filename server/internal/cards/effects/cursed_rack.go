package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cursed Rack — Artifact {4}:
//
//	"As this artifact enters, choose an opponent.
//	 The chosen player's maximum hand size is four."
//
// The opponent is chosen as it enters and stored on the permanent
// (ChoosePlayerAsEnters, Card.ChosenPlayer), so the choice survives a
// change of controller (the 2004-10-04 ruling). The static reaches that
// player only (game.HandSizeChosenPlayer, ADR 0113 §3, #2074) and is
// folded in CR 613.11 timestamp order: a Null Profusion then a Cursed
// Rack is four, the other order is two (the 2009-10-01 ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f04603f-8f91-405b-b6ac-d2b66f05e32f",
		Name:         "Cursed Rack",
		Completeness: CompletenessFull,
		AsEnters:     ChoosePlayerAsEnters("Cursed Rack", Opponents),
		HandSize: []game.HandSizeStatic{{
			Players: game.HandSizeChosenPlayer,
			Kind:    game.HandSizeSet,
			N:       4,
		}},
	})
}

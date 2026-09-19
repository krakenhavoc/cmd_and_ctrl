package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virus Beetle — 1/1 Artifact Creature — Insect for {1}{B} (EDHREC
// rank 4039):
//
//	"When this creature enters, each opponent discards a card."
//
// Burglar Rat's artifact cousin: the same ETB on the same body for the
// same price, but an ARTIFACT creature, which is why the affinity and
// Karn lists run it alongside rather than instead. It is in the batch
// for exactly that overlap — the discard body is shared
// (eachOpponentDiscardsOne), so the only thing this card adds is a
// second caller, and a second caller is how a shared body stays
// honest.
//
// One prompt per opponent, addressed to that opponent and offering
// only their own hand, so nobody discards on anyone else's behalf. An
// opponent with an empty hand is skipped rather than stalling the
// trigger. Each discard is a real discard, so the discard payoffs
// (Megrim, Waste Not) see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1e489f6-37e5-469e-bca3-0e9d3fc3fa95",
		Name:         "Virus Beetle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Virus Beetle — each opponent discards a card", eachOpponentDiscardsOne),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brago's Representative — Creature — Human Advisor {2}{W}, 1/4 (EDHREC
// rank 11978):
//
//	"While voting, you get an additional vote. (The votes can be for
//	 different choices or for the same choice.)"
//
// CR 701.38d (ADR 0146): its controller casts one more ballot in every
// vote, at the moment they would vote, and must cast it. Two of them
// are two more. It does nothing for a vote while it is not on the
// battlefield, or once it has lost its abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8877b896-8458-4828-9ff6-9e0ecc180cf7",
		Name:         "Brago's Representative",
		Completeness: CompletenessFull,
		ExtraVote:    game.ExtraVoteYouGet,
	})
}

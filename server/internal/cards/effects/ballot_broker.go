package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ballot Broker — Creature — Human Advisor {2}{W}, 2/3 (EDHREC rank
// 11879):
//
//	"While voting, you may vote an additional time. (The votes can be
//	 for different choices or for the same choice.)"
//
// CR 701.38d (ADR 0146): after its controller's own vote they are
// offered one more ballot, with "Don't vote again" beside the options.
// It does nothing for a vote while it is not on the battlefield, or
// once it has lost its abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8df915f-f509-4c6c-9261-a5a16c3f06c5",
		Name:         "Ballot Broker",
		Completeness: CompletenessFull,
		ExtraVote:    game.ExtraVoteYouMay,
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Valeyard — Legendary Creature — Time Lord Noble {2}{U}{B}{R}, 4/5:
//
//	"If an opponent would face a villainous choice, they face that
//	 choice an additional time. (They can make the same or different
//	 choices.)
//	 While voting, you may vote an additional time."
//
// The second ability is ADR 0146's extra vote (game.ExtraVoteYouMay,
// CR 701.38d), the one Ballot Broker and Tivit print: after the
// Valeyard's controller has cast their votes in a vote, they are offered
// one more, and may decline it.
//
// The first ability is CR 701.55c's replacement effect (#3140): "A
// replacement effect may replace an instruction to face a villainous
// choice with an instruction to face that choice some number of
// additional times. In that case, the entire process … is performed for
// that player the appropriate number of times one at a time." It is
// Spec.OpponentsFaceVillainousChoicesAgain, a count read when a choice
// is faced (Game.VillainousChoiceTimesForEffect): each Valeyard under
// an opponent's control adds one question, "the same or different
// choices" each time, so two of them make three. It applies to the
// players who are its controller's opponents, not to its controller.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ed202f1d-3511-45f4-bfbc-e6b76d4311d1",
		Name:         "The Valeyard",
		Completeness: CompletenessFull,
		ExtraVote:    game.ExtraVoteYouMay,

		OpponentsFaceVillainousChoicesAgain: true,
	})
}

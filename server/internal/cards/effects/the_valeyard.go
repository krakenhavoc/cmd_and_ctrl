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
// The first ability is NOT implemented. A villainous choice (CR 701.55)
// is a keyword action with no engine support at all: nothing in this
// game asks one, so there is no choice for an opponent to face, and no
// card in the catalog that could make one. Declared weaker than printed,
// which is the safe direction: the first ability never has a choice to
// double, so it cannot change a game today. It is its own seam, with the
// keyword action itself, and not a flag on this card.
func init() {
	Register(Spec{
		OracleID:     "ed202f1d-3511-45f4-bfbc-e6b76d4311d1",
		Name:         "The Valeyard",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Villainous choices aren't in the game yet, so its first ability does nothing."},
		ExtraVote:    game.ExtraVoteYouMay,
	})
}

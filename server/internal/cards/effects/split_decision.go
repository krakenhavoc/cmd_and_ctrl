package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Split Decision — Instant {1}{U}:
//
//	"Will of the council — Choose target instant or sorcery spell.
//	 Starting with you, each player votes for denial or duplication. If
//	 denial gets more votes, counter the spell. If duplication gets more
//	 votes or the vote is tied, copy the spell. You may choose new
//	 targets for the copy."
//
// The target is chosen as Split Decision is cast (CR 115.1) and checked
// again as it resolves (CR 608.2b): a spell that left the stack means
// no vote at all. The vote is ADR 0146's (CR 701.38); the stack is
// frozen while the players vote, so the spell is still there when they
// finish. A tie goes to duplication, as printed: the copy is Split
// Decision's controller's (CR 707.10), and they may choose new targets
// for it.
//
// A bot votes duplication whoever cast it: the hint cannot see whose
// spell is the target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f725d51-e133-488c-9ca9-1c7b0523b5bd",
		Name:         "Split Decision",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			return Vote{
				Question:      "Split Decision — vote for denial or duplication",
				Words:         []string{"denial", "duplication"},
				ForController: []int{1, 2},
				ForOpponents:  []int{1, 2},
				Then:          splitDecisionVoted,
				Carry:         []uuid.UUID{legal[0].ID},
			}.Apply(ctx)
		},
	})
}

var splitDecisionVoted = VoteResultThen("vote/split-decision", func(ctx *Context, r game.VoteResult) error {
	if len(r.Carry) == 0 {
		return nil
	}
	spell := r.Carry[0]
	if r.MoreVotes(0, 1) {
		return CounterTarget{StackID: spell}.Apply(ctx)
	}
	return CopySpell{StackID: spell, Controller: ctx.Controller(), ChooseNewTargets: true}.Apply(ctx)
})

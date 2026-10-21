package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lieutenants of the Guard — Creature — Human Soldier {4}{W}, 2/2:
//
//	"Council's dilemma — When this creature enters, starting with you,
//	 each player votes for strength or numbers. Put a +1/+1 counter on
//	 this creature for each strength vote and create a 1/1 white
//	 Soldier creature token for each numbers vote."
//
// An enters trigger whose resolution is ADR 0146's council's dilemma
// (CR 701.38): every ballot pays its controller, a counter per strength
// vote and a Soldier per numbers vote. "This creature" is the object
// that entered (CR 400.7): one that left while the trigger waited gets
// no counters, and the Soldiers are made regardless.
//
// A bot votes numbers for its own Lieutenants (bodies spread out) and
// strength for an opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80745dea-2335-43fc-9915-f62e0436e73b",
		Name:         "Lieutenants of the Guard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lieutenants of the Guard — each player votes for strength or numbers", voteForThisCreature(Vote{
				Question:      "Lieutenants of the Guard — vote for strength or numbers",
				Words:         []string{"strength", "numbers"},
				ForController: []int{1, 2},
				ForOpponents:  []int{2, 1},
				Then:          lieutenantsOfTheGuardVoted,
			})),
		},
	})
}

const lieutenantsOfTheGuardRecord = "lieutenants-of-the-guard/votes"

var lieutenantsOfTheGuardVoted = VoteResultThen("vote/lieutenants-of-the-guard", func(ctx *Context, r game.VoteResult) error {
	recordVotedCreature(ctx, lieutenantsOfTheGuardRecord, r)
	return lieutenantsOfTheGuardSteps.Run(ctx)
})

// lieutenantsOfTheGuardSteps is the counters and then the Soldiers: the tokens
// wait for a prompt the counters ask (a CR 616 order; CR 608.2c, #4155).
var lieutenantsOfTheGuardSteps = RegisterThenSteps("then-steps/lieutenants-of-the-guard-votes",
	Instruction(func(ctx *Context) error {
		return countersOnRecordedCreature(ctx, lieutenantsOfTheGuardRecord, 0)
	}),
	Instruction(func(ctx *Context) error {
		if n := recordedTally(ctx, lieutenantsOfTheGuardRecord, 1); n > 0 {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Soldier"), N: n}.Apply(ctx)
		}
		return nil
	}),
)

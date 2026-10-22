package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orchard Elemental — Creature — Elemental {5}{G}, 2/2:
//
//	"Council's dilemma — When this creature enters, starting with you,
//	 each player votes for sprout or harvest. Put two +1/+1 counters on
//	 this creature for each sprout vote. You gain 3 life for each
//	 harvest vote."
//
// An enters trigger whose resolution is ADR 0146's council's dilemma
// (CR 701.38): every ballot pays its controller, two counters per
// sprout vote and 3 life per harvest vote. "This creature" is the
// object that entered (CR 400.7): one that left while the trigger
// waited gets no counters, and the life is gained regardless.
//
// A bot votes sprout for its own Elemental and harvest for an
// opponent's, the smaller gift.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "189b7c2b-81df-41c3-94ae-6b32dbd48cec",
		Name:         "Orchard Elemental",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Orchard Elemental — each player votes for sprout or harvest", voteForThisCreature(Vote{
				Question:      "Orchard Elemental — vote for sprout or harvest",
				Words:         []string{"sprout", "harvest"},
				ForController: []int{2, 1},
				ForOpponents:  []int{0, 1},
				Then:          orchardElementalVoted,
			})),
		},
	})
}

// orchardElementalRecord is the creature voted for and both tallies.
const orchardElementalRecord = "orchard-elemental/votes"

var orchardElementalVoted = VoteResultThen("vote/orchard-elemental", func(ctx *Context, r game.VoteResult) error {
	recordVotedCreature(ctx, orchardElementalRecord, r)
	return orchardElementalSteps.Run(ctx)
})

// orchardElementalSteps is the counters and then the life: the life waits for a
// prompt the counters ask (a CR 616 order; CR 608.2c, #4155).
var orchardElementalSteps = RegisterThenSteps("then-steps/orchard-elemental-votes",
	Instruction(func(ctx *Context) error {
		r := ctx.StepRecord(orchardElementalRecord)
		if len(r.IDs) == 0 || len(r.Ns) < 2 || r.IDs[0] == uuid.Nil || r.Ns[0] <= 0 {
			return nil
		}
		return AddCounter{Target: r.IDs[0], Kind: "+1/+1", N: 2 * r.Ns[0]}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		return GainLife{Amount: 3 * recordedTally(ctx, orchardElementalRecord, 1)}.Apply(ctx)
	}),
)

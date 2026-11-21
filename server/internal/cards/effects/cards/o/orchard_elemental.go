package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "189b7c2b-81df-41c3-94ae-6b32dbd48cec",
		Name:         "Orchard Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Orchard Elemental — each player votes for sprout or harvest", effects.VoteForThisCreature(effects.Vote{
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

var orchardElementalVoted = effects.VoteResultThen("vote/orchard-elemental", func(ctx *effects.Context, r game.VoteResult) error {
	effects.RecordVotedCreature(ctx, orchardElementalRecord, r)
	return orchardElementalSteps.Run(ctx)
})

// orchardElementalSteps is the counters and then the life: the life waits for a
// prompt the counters ask (a CR 616 order; CR 608.2c, #4155).
var orchardElementalSteps = effects.RegisterThenSteps("then-steps/orchard-elemental-votes",
	effects.Instruction(func(ctx *effects.Context) error {
		r := ctx.StepRecord(orchardElementalRecord)
		if len(r.IDs) == 0 || len(r.Ns) < 2 || r.IDs[0] == uuid.Nil || r.Ns[0] <= 0 {
			return nil
		}
		return effects.AddCounter{Target: r.IDs[0], Kind: "+1/+1", N: 2 * r.Ns[0]}.Apply(ctx)
	}),
	effects.Instruction(func(ctx *effects.Context) error {
		return effects.GainLife{Amount: 3 * effects.RecordedTally(ctx, orchardElementalRecord, 1)}.Apply(ctx)
	}),
)

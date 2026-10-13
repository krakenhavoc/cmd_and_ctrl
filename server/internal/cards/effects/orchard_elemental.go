package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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

var orchardElementalVoted = VoteResultThen("vote/orchard-elemental", func(ctx *Context, r game.VoteResult) error {
	if err := countersOnVotedCreature(ctx, r, 2*r.Votes(0)); err != nil {
		return err
	}
	return GainLife{Amount: 3 * r.Votes(1)}.Apply(ctx)
})

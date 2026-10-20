package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emissary Green — Legendary Creature — Human Advisor {4}{G}, 3/3:
//
//	"Whenever Emissary Green attacks, starting with you, each player
//	 votes for profit or security. You create a number of Treasure
//	 tokens equal to twice the number of profit votes. Put a number of
//	 +1/+1 counters on each creature you control equal to the number of
//	 security votes."
//
// An attack trigger whose resolution is ADR 0146's vote (CR 701.38):
// every ballot pays Emissary Green's controller, two Treasures per
// profit vote and a counter on every creature they control per security
// vote. The creatures are read as the vote ends, so one that entered
// while the players voted is included and one that left is not. The
// counters go on all of them at once, Emissary Green included if it is
// still there.
//
// A bot votes profit for its own Emissary Green and security for an
// opponent's, the smaller gift.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c43e7fa5-ff93-42d9-af79-ee417a8e41a7",
		Name:         "Emissary Green",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Emissary Green — each player votes for profit or security", Do(Vote{
				Question:      "Emissary Green — vote for profit or security",
				Words:         []string{"profit", "security"},
				ForController: []int{2, 1},
				ForOpponents:  []int{0, 1},
				Then:          emissaryGreenVoted,
			})),
		},
	})
}

var emissaryGreenVoted = VoteResultThen("vote/emissary-green", func(ctx *Context, r game.VoteResult) error {
	ctx.RecordStep(emissaryGreenVotesRecord, nil, r.Votes(0), r.Votes(1))
	return emissaryGreenRewards.Run(ctx)
})

// emissaryGreenVotesRecord is the votes for profit and for security.
const emissaryGreenVotesRecord = "emissary-green/votes"

// emissaryGreenRewards is the Treasures and then the counters: the counters
// wait for a prompt the Treasures' creation asks (a CR 616 order on tokens; CR
// 608.2c, #4155).
var emissaryGreenRewards = RegisterThenSteps("then-steps/emissary-green-rewards",
	Instruction(func(ctx *Context) error {
		if n := 2 * ctx.StepRecord(emissaryGreenVotesRecord).Ns[0]; n > 0 {
			return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: n}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		n := ctx.StepRecord(emissaryGreenVotesRecord).Ns[1]
		if n == 0 {
			return nil
		}
		var mine []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.Controller == ctx.Controller() && c.IsCreature() {
				mine = append(mine, c.InstanceID)
			}
		}
		group := ctx.asGroupMember()
		for _, id := range mine {
			if err := (AddCounter{Target: id, Kind: "+1/+1", N: n}).Apply(group); err != nil {
				return err
			}
		}
		return nil
	}),
)

package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Truth or Consequences — Sorcery {2}{U}{R}:
//
//	"Secret council — Each player secretly votes for truth or
//	 consequences, then those votes are revealed. You draw cards equal
//	 to the number of truth votes. Then choose an opponent at random.
//	 Truth or Consequences deals 3 damage to that player for each
//	 consequences vote."
//
// A secret vote (ADR 0180): nobody sees a ballot until the last one is
// in, then the log reveals them together. The draw is one instruction
// of N cards; the opponent is drawn from the keyed random stream
// (ChooseAtRandomForEffect), and the damage is one instance of 3 per
// consequences vote.
//
// A bot casting it votes consequences (both words help it, and damage
// is the bigger swing); an opponent's bot votes truth, which only
// draws the caster cards, over a vote that may burn it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58d944fb-d1ba-4141-baaf-37dd10978656",
		Name:         "Truth or Consequences",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Truth or Consequences — secretly vote for truth or consequences",
				Words:         []string{"truth", "consequences"},
				Secret:        true,
				ForController: []int{1, 2},
				ForOpponents:  []int{1, 0},
				Then:          truthOrConsequencesVoted,
			}.Apply(ctx)
		},
	})
}

var truthOrConsequencesVoted = VoteResultThen("vote/truth-or-consequences", func(ctx *Context, r game.VoteResult) error {
	if err := (DrawCards{N: r.Votes(0)}).Apply(ctx); err != nil {
		return err
	}
	opponents := ctx.Opponents()
	if len(opponents) == 0 {
		return nil
	}
	pick := ctx.Game.ChooseAtRandomForEffect(randomDraw(ctx), opponents, 1)
	if n := r.Votes(1); n > 0 && len(pick) == 1 {
		return ctx.Game.DealDamageToPlayerForEffect(ctx.Source(), pick[0], 3*n)
	}
	return nil
})

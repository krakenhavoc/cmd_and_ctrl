package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Council's Judgment — Sorcery {1}{W}{W} (EDHREC rank 3537):
//
//	"Will of the council — Starting with you, each player votes for a
//	 nonland permanent you don't control. Exile each permanent with the
//	 most votes or tied for most votes."
//
// An object vote (CR 701.38b, ADR 0146). The options are the nonland
// permanents the caster doesn't control as the spell resolves; nothing
// is targeted, so hexproof, shroud and protection don't stop a
// permanent being voted for or exiled, which is the card's point. Each
// player votes for one of them (an opponent may vote for their own),
// and every permanent with the most votes, ties included, is exiled at
// once. With nothing to vote for, nothing happens.
//
// A bot votes for the most valuable permanent its own seat doesn't
// control (aiseat/heuristic/vote.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acf4d685-a718-42f7-9a7a-dee069ae8db2",
		Name:         "Council's Judgment",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.Battlefield.Cards {
				if c.Controller != ctx.Controller() && !c.IsLand() {
					ids = append(ids, c.InstanceID)
				}
			}
			return Vote{
				Question: "Council's Judgment — vote for a nonland permanent its caster doesn't control",
				Options:  VoteForPermanents(ctx.Game, ids),
				Then:     councilsJudgmentVoted,
			}.Apply(ctx)
		},
	})
}

var councilsJudgmentVoted = VoteResultThen("vote/councils-judgment", func(ctx *Context, r game.VoteResult) error {
	var exile []uuid.UUID
	for _, i := range r.MostVotes() {
		if id := VotedPermanent(r, i); id != uuid.Nil && onBattlefield(ctx.Game, id) {
			exile = append(exile, id)
		}
	}
	if len(exile) > 0 {
		ctx.Game.ExileCardsForEffect(exile)
	}
	return nil
})

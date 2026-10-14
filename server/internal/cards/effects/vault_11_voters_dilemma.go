package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vault 11: Voter's Dilemma — Enchantment — Saga {2}{W}{B}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — For each opponent, you create a 1/1 white Human Soldier
//	     creature token.
//	 II, III — Each player secretly votes for up to one creature, then
//	     those votes are revealed. If no creature got votes, each player
//	     draws a card. Otherwise, destroy each creature with the most
//	     votes or tied for most votes."
//
// Chapters II and III are a secret object vote (ADR 0180) over every
// creature on the battlefield as the chapter resolves, with a vote for
// nothing on every ballot ("up to one creature", Vote.Abstain). If no
// creature got a vote (everyone voted for nothing, or there was no
// creature), each player draws a card in turn order from the active
// player; otherwise every creature with the most votes or tied, still
// on the battlefield, is destroyed at once. Like Trial of a Time Lord,
// the Saga is not sacrificed while its last chapter's vote is open.
//
// No simplification.
func init() {
	const name = "Vault 11: Voter's Dilemma"
	vote := func(n int) game.TriggeredAbility {
		return ChapterTrigger(n, SagaChapterLabel(name, n, "each player secretly votes for up to one creature"),
			func(g *game.Game, item *game.StackItem) error {
				var ids []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCreature() {
						ids = append(ids, c.InstanceID)
					}
				}
				return Vote{
					Question: name + " — secretly vote for up to one creature",
					Options:  VoteForPermanents(g, ids),
					Secret:   true,
					Abstain:  "No creature",
					Then:     vault11Voted,
				}.Apply(NewContext(g, item))
			})
	}
	Register(Spec{
		OracleID:     "e2c00aff-696c-4b20-acd4-95b71b5f8b10",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "for each opponent, create a 1/1 white Human Soldier creature token"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := len(ctx.Opponents())
					if n == 0 {
						return nil
					}
					return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 white Human Soldier"), N: n}.Apply(ctx)
				}),
			vote(2),
			vote(3),
		},
	})
}

var vault11Voted = VoteResultThen("vote/vault-11-voters-dilemma", func(ctx *Context, r game.VoteResult) error {
	most := r.MostVotes()
	if len(most) == 0 {
		for _, p := range apnapLivingSeats(ctx.Game) {
			if err := ctx.Game.DrawNForEffect(p, 1); err != nil {
				return err
			}
		}
		return nil
	}
	var doomed []uuid.UUID
	for _, i := range most {
		if id := VotedPermanent(r, i); id != uuid.Nil && onBattlefield(ctx.Game, id) {
			doomed = append(doomed, id)
		}
	}
	if len(doomed) > 0 {
		ctx.Game.DestroyPermanentsForEffect(doomed)
	}
	return nil
})

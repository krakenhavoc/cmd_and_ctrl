package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coercive Portal — Artifact {4} (EDHREC rank 8853):
//
//	"Will of the council — At the beginning of your upkeep, starting
//	 with you, each player votes for carnage or homage. If carnage gets
//	 more votes, sacrifice this artifact and destroy all nonland
//	 permanents. If homage gets more votes or the vote is tied, draw a
//	 card."
//
// An upkeep trigger whose resolution is ADR 0146's vote (CR 701.38).
// Carnage sacrifices the Portal first and then destroys every nonland
// permanent. "This artifact" is the object the trigger came from, so a
// Portal that left the battlefield, or came back as a new object, while
// the trigger waited is not sacrificed (CR 400.7), and the board is
// destroyed all the same. That is read as the trigger resolves and
// carried on the vote, since nothing can move the Portal while the
// players vote. A tie goes to homage, as printed.
//
// A bot voting on its own Portal votes homage (a card, and it keeps its
// board); an opponent's bot votes homage too, since carnage takes its
// own permanents with everyone else's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eea4feef-2595-4ed2-92a0-11ccb3e7d40e",
		Name:         "Coercive Portal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Coercive Portal — each player votes for carnage or homage", coercivePortalVote),
		},
	})
}

func coercivePortalVote(g *game.Game, item *game.StackItem) error {
	var portal []uuid.UUID
	if onBattlefield(g, item.SourceCardID) && !sourceIsNewObject(g, item) {
		portal = []uuid.UUID{item.SourceCardID}
	}
	return Vote{
		Question:      "Coercive Portal — vote for carnage or homage",
		Words:         []string{"carnage", "homage"},
		ForController: []int{0, 1},
		ForOpponents:  []int{0, 1},
		Then:          coercivePortalVoted,
		Carry:         portal,
	}.Apply(NewContext(g, item))
}

var coercivePortalVoted = VoteResultThen("vote/coercive-portal", func(ctx *Context, r game.VoteResult) error {
	if !r.MoreVotes(0, 1) {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}
	item := ctx.Item
	destroy := func(g *game.Game, _ []uuid.UUID) error {
		return DestroyAllMatching{Match: Nonland()}.Apply(NewContext(g, item))
	}
	if len(r.Carry) == 0 || !onBattlefield(ctx.Game, r.Carry[0]) {
		return destroy(ctx.Game, nil)
	}
	return ctx.Game.SacrificeAllThenForEffect(r.Source, r.Carry[:1], destroy)
})

package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "eea4feef-2595-4ed2-92a0-11ccb3e7d40e",
		Name:         "Coercive Portal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			// ADR 0126 §6: the sweep runs only when carnage wins the vote.
			effects.TriggerWithPurpose(effects.AtYourUpkeep("Coercive Portal — each player votes for carnage or homage", coercivePortalVote),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}}),
		},
	})
}

func coercivePortalVote(g *game.Game, item *game.StackItem) error {
	var portal []uuid.UUID
	if effects.OnBattlefield(g, item.SourceCardID) && !effects.SourceIsNewObject(g, item) {
		portal = []uuid.UUID{item.SourceCardID}
	}
	return effects.Vote{
		Question:      "Coercive Portal — vote for carnage or homage",
		Words:         []string{"carnage", "homage"},
		ForController: []int{0, 1},
		ForOpponents:  []int{0, 1},
		Then:          coercivePortalVoted,
		Carry:         portal,
	}.Apply(effects.NewContext(g, item))
}

var coercivePortalVoted = effects.VoteResultThen("vote/coercive-portal", func(ctx *effects.Context, r game.VoteResult) error {
	if !r.MoreVotes(0, 1) {
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}
	if len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return effects.DestroyAllMatching{Match: effects.Nonland()}.Apply(ctx)
	}
	return ctx.Game.SacrificeAllKeyedForEffect(r.Source, r.Carry[:1], ctx.ItemTail(coercivePortalCarnage))
})

// coercivePortalCarnage is carnage's "then destroy all nonland permanents",
// after the sacrifice, keyed (#4253 slice 10a).
var coercivePortalCarnage = effects.CardTail("then/coercive-portal-carnage", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DestroyAllMatching{Match: effects.Nonland()}.Apply(ctx)
})

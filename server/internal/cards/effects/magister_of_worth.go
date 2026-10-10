package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magister of Worth — Creature — Angel {4}{W}{B}, 4/4 (EDHREC rank
// 9867):
//
//	"Flying
//	 Will of the council — When this creature enters, starting with
//	 you, each player votes for grace or condemnation. If grace gets
//	 more votes, each player returns each creature card from their
//	 graveyard to the battlefield. If condemnation gets more votes or
//	 the vote is tied, destroy all creatures other than this creature."
//
// An enters trigger whose resolution is ADR 0146's vote (CR 701.38).
// Grace returns every creature card from every graveyard at once, each
// under its owner's control (ReturnFromGraveyardTogether, so each sees
// the others enter). Condemnation, or a tie, destroys every creature
// except the Magister. "This creature" is the object that entered: one
// that has left, or come back as a new object, while the trigger
// waited is destroyed with the rest (CR 400.7). It is read as the
// trigger resolves and carried on the vote.
//
// The bot hint is read off the graveyards as the vote starts: grace
// helps whoever has more creature cards in their graveyard, so the
// Magister's controller votes grace when they have at least as many as
// the opponents put together, and condemnation otherwise; an opponent's
// bot votes the other way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0429d146-332f-4fb0-8273-c73679ca6d47",
		Name:            "Magister of Worth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Magister of Worth — each player votes for grace or condemnation", magisterOfWorthVote),
		},
	})
}

func magisterOfWorthVote(g *game.Game, item *game.StackItem) error {
	var self []uuid.UUID
	if onBattlefield(g, item.SourceCardID) && !sourceIsNewObject(g, item) {
		self = []uuid.UUID{item.SourceCardID}
	}
	mine, theirs := 0, 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if !c.IsCreature() {
				continue
			}
			if p.ID == item.Controller {
				mine++
			} else {
				theirs++
			}
		}
	}
	forController, forOpponents := []int{0, 1}, []int{1, 0}
	if mine >= theirs {
		forController, forOpponents = []int{1, 0}, []int{0, 1}
	}
	return Vote{
		Question:      "Magister of Worth — vote for grace or condemnation",
		Words:         []string{"grace", "condemnation"},
		ForController: forController,
		ForOpponents:  forOpponents,
		Then:          magisterOfWorthVoted,
		Carry:         self,
	}.Apply(NewContext(g, item))
}

var magisterOfWorthVoted = VoteResultThen("vote/magister-of-worth", func(ctx *Context, r game.VoteResult) error {
	if r.MoreVotes(0, 1) {
		var dead []uuid.UUID
		for _, p := range ctx.Game.Seats {
			if p == nil || p.Graveyard == nil {
				continue
			}
			for _, c := range p.Graveyard.Cards {
				if c.IsCreature() {
					dead = append(dead, c.InstanceID)
				}
			}
		}
		return ReturnFromGraveyardTogether{Targets: dead}.Apply(ctx)
	}
	spared := uuid.Nil
	if len(r.Carry) > 0 {
		spared = r.Carry[0]
	}
	return DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.InstanceID != spared
	}}.Apply(ctx)
})

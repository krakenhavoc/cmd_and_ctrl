package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expropriate — Sorcery {7}{U}{U} (EDHREC rank 2487):
//
//	"Council's dilemma — Starting with you, each player votes for time
//	 or money. For each time vote, take an extra turn after this one.
//	 For each money vote, choose a permanent owned by the voter and
//	 gain control of it. Exile Expropriate."
//
// Council's dilemma is a vote whose every ballot does something (ADR
// 0146, CR 701.38). The caster takes one extra turn per time vote. For
// each money vote the caster chooses a permanent the voter OWNS,
// whoever controls it now, and gains control of it for good (CR
// 611.2a: no duration is printed). A voter who voted money twice gives
// up two different permanents; one who owns fewer permanents than money
// votes gives up all of them. The choices are one prompt per voter, in
// the order they voted.
//
// "Exile Expropriate" is the spell's last instruction. The spell exiles
// itself as it starts the vote rather than after it: the players vote
// while it is still resolving, nothing can see it in between, and so it
// never reaches its owner's graveyard (#489's spellMovedItselfLocked
// keeps the resolution from burying it).
//
// A bot casting it votes time; an opponent's bot votes time too, since
// money hands the caster its own best permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a69265b2-0e37-4d86-866e-e4a923233b4d",
		Name:         "Expropriate",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (Vote{
				Question:      "Expropriate — vote for time or money",
				Words:         []string{"time", "money"},
				ForController: []int{2, 1},
				ForOpponents:  []int{1, 0},
				Then:          expropriateVoted,
			}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

var expropriateVoted = VoteResultThen("vote/expropriate", func(ctx *Context, r game.VoteResult) error {
	if n := r.Votes(0); n > 0 {
		if err := (TakeExtraTurn{N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	owed := map[uuid.UUID]int{}
	var voters []uuid.UUID
	for _, v := range r.VotersFor(1) {
		if owed[v] == 0 {
			voters = append(voters, v)
		}
		owed[v]++
	}
	if len(voters) == 0 {
		return nil
	}
	return ChoosePermanents{
		Of:       voters,
		Question: "Expropriate — choose a permanent owned by the voter for each of their money votes, and gain control of it",
		Candidates: func(g *game.Game, voter uuid.UUID) ([]uuid.UUID, int, int) {
			var ids []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Owner == voter {
					ids = append(ids, c.InstanceID)
				}
			}
			n := owed[voter]
			return ids, n, n
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			for _, id := range picked.Cards() {
				if err := (GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: "Expropriate — gain control"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(ctx)
})

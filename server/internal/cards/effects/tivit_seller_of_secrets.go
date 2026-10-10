package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tivit, Seller of Secrets — Legendary Creature — Sphinx Rogue
// {3}{W}{U}{B}, 6/6 (EDHREC rank 4475):
//
//	"Flying, ward {3}
//	 Council's dilemma — Whenever Tivit enters or deals combat damage to
//	 a player, starting with you, each player votes for evidence or
//	 bribery. For each evidence vote, investigate. For each bribery
//	 vote, create a Treasure token.
//	 While voting, you may vote an additional time. (The votes can be
//	 for different choices or for the same choice.)"
//
// One trigger with two conditions (Aerial Extortionist's), whose
// resolution is ADR 0146's vote (CR 701.38). Every ballot pays Tivit's
// controller: a Clue per evidence vote, a Treasure per bribery vote.
// The last line is the ExtraVote static (CR 701.38d), so Tivit's
// controller is offered a second ballot in every vote while Tivit is on
// the battlefield, its own included.
//
// A bot votes bribery for its own Tivit (mana now) and evidence for an
// opponent's (a Clue costs {2} more to use).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f4e0b557-4f80-4b7c-bfcc-6d4c468faf2e",
		Name:            "Tivit, Seller of Secrets",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ExtraVote:       game.ExtraVoteYouMay,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Tivit, Seller of Secrets — ward {3}"),
			{
				Watches: []game.EventKind{game.EventETB, game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b26SelfEnteredOrDealtCombatDamageToPlayer(ev, source, g)
				},
				Key: "Tivit, Seller of Secrets — each player votes for evidence or bribery",
				Effect: Do(Vote{
					Question:      "Tivit, Seller of Secrets — vote for evidence or bribery",
					Words:         []string{"evidence", "bribery"},
					ForController: []int{1, 2},
					ForOpponents:  []int{1, 0},
					Then:          tivitVoted,
				}),
			},
		},
	})
}

var tivitVoted = VoteResultThen("vote/tivit-seller-of-secrets", func(ctx *Context, r game.VoteResult) error {
	if n := r.Votes(0); n > 0 {
		if err := (CreateToken{Controller: ctx.Controller(), Template: ClueToken(), N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	if n := r.Votes(1); n > 0 {
		return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: n}.Apply(ctx)
	}
	return nil
})

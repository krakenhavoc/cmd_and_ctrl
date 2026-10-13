package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Travel Through Caradhras — Sorcery {5}{G}:
//
//	"Council's dilemma — Starting with you, each player votes for
//	 Redhorn Pass or Mines of Moria. For each Redhorn Pass vote, search
//	 your library for a basic land card and put it onto the battlefield
//	 tapped. If you search your library this way, shuffle. For each
//	 Mines of Moria vote, return a card from your graveyard to your
//	 hand.
//	 Exile Travel Through Caradhras."
//
// A council's dilemma (ADR 0146, CR 701.38): every ballot pays the
// caster. The Redhorn Pass votes are one search for that many basic
// lands, all put onto the battlefield tapped, and one shuffle at the
// end; the Mines of Moria votes are one choice of that many cards from
// the caster's graveyard (fewer if it holds fewer), returned together.
// Searching once for N cards is the same result as N searches in a
// row: the only difference is a shuffle between them nobody can see.
//
// "Exile Travel Through Caradhras" is the spell's last instruction. The
// spell exiles itself as it starts the vote, as Expropriate does: the
// players vote while it is still resolving, so it never reaches the
// graveyard and cannot be one of the cards Mines of Moria returns.
//
// A bot casting it votes Redhorn Pass; an opponent's bot votes Mines
// of Moria, the smaller gift.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d47a98b-5700-4096-b080-32df4d79910c",
		Name:         "Travel Through Caradhras",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (Vote{
				Question:      "Travel Through Caradhras — vote for Redhorn Pass or Mines of Moria",
				Words:         []string{"Redhorn Pass", "Mines of Moria"},
				ForController: []int{2, 1},
				ForOpponents:  []int{0, 1},
				Then:          travelThroughCaradhrasVoted,
			}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

var travelThroughCaradhrasVoted = VoteResultThen("vote/travel-through-caradhras", func(ctx *Context, r game.VoteResult) error {
	if n := r.Votes(0); n > 0 {
		if err := (SearchLibrary{
			Player:        ctx.Controller(),
			Predicate:     IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         n,
			TappedOnEntry: true,
			Shuffle:       true,
			Reason:        "Travel Through Caradhras — search for a basic land card for each Redhorn Pass vote",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if n := r.Votes(1); n > 0 {
		return ReturnChosenFromGraveyard{
			Question: "Travel Through Caradhras — return a card from your graveyard to your hand for each Mines of Moria vote",
			Except:   r.Source,
			Min:      n,
			Max:      n,
			ToHand:   true,
		}.Apply(ctx)
	}
	return nil
})

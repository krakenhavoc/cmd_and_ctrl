package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grudge Keeper — Creature — Zombie Wizard {1}{B}, 2/1:
//
//	"Whenever players finish voting, each opponent who voted for a choice
//	 you didn't vote for loses 2 life."
//
// ADR 0146's amendment of 2026-10-13: the trigger is
// WheneverPlayersFinishVoting, and "a choice you didn't vote for" is
// read off the finished vote's ballots as the ability resolves
// (finish_voting.go). It fires for every vote at the table, including
// one an opponent called; a Grudge Keeper's controller who cast no
// ballot (they left the vote, or had no vote) makes every choice an
// opponent voted for one "you didn't vote for".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffbf8459-fda2-4c1d-9841-c001c37d2141",
		Name:         "Grudge Keeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverPlayersFinishVoting("Grudge Keeper — each opponent who voted for a choice you didn't vote for loses 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, opp := range OpponentsWhoVotedForAChoiceYouDidntVoteFor(ctx) {
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -2); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

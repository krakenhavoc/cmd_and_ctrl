package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erestor of the Council — Legendary Creature — Elf Noble {1}{G}{U}:
//
//	"Whenever players finish voting, each opponent who voted for a choice
//	 you voted for creates a Treasure token. You scry X, where X is the
//	 number of opponents who voted for a choice you didn't vote for. Draw
//	 a card."
//
// ADR 0146's amendment of 2026-10-13 (finish_voting.go). The Treasures go
// to the opponents who share a choice, each under their own control; X
// counts opponents, not ballots, so an opponent who voted twice for
// choices you didn't is one. Both groups are read off the finished vote's
// ballots as the ability resolves, and an opponent who voted for a choice
// you voted for AND one you didn't is in both.
//
// "Draw a card" is the last sentence, so it rides Scry.Then: the scry is
// a prompt, and a draw written after it would take a card the player is
// still deciding about. A scry of zero (nobody voted apart from you)
// still draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b65e1dff-9423-4840-a6af-78cb3e6f4f31",
		Name:         "Erestor of the Council",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WheneverPlayersFinishVoting("Erestor of the Council — Treasures for those who voted with you, scry for those who didn't, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, opp := range OpponentsWhoVotedForAChoiceYouVotedFor(ctx) {
						if err := (CreateToken{Controller: opp, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					me := ctx.Controller()
					x := len(OpponentsWhoVotedForAChoiceYouDidntVoteFor(ctx))
					return Scry{
						Player: me,
						N:      x,
						Then:   func(g *game.Game) error { return g.DrawNForEffect(me, 1) },
					}.Apply(ctx)
				}),
		},
	})
}

package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b65e1dff-9423-4840-a6af-78cb3e6f4f31",
		Name:         "Erestor of the Council",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverPlayersFinishVoting("Erestor of the Council — Treasures for those who voted with you, scry for those who didn't, draw a card",
				// The scry waits for a prompt a Treasure's creation asks (a CR 616 order; CR 608.2c, #4155).
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						for _, opp := range effects.OpponentsWhoVotedForAChoiceYouVotedFor(ctx) {
							if err := (effects.CreateToken{Controller: opp, Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						me := ctx.Controller()
						x := len(effects.OpponentsWhoVotedForAChoiceYouDidntVoteFor(ctx))
						return effects.Scry{
							Player:   me,
							N:        x,
							ThenTail: ctx.TailN(effects.ThenDrawCards, 1),
						}.Apply(ctx)
					}),
				)),
		},
	})
}

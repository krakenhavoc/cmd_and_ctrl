package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illusion of Choice — Instant {U}:
//
//	"You choose how each player votes this turn.
//	 Draw a card."
//
// A turn-scoped rule on ADR 0146's vote (CR 701.38, amendment of
// 2026-10-13): a stored ModVoteChooser record naming the caster, swept
// at cleanup (CR 514.2) and carried by the snapshot like every scoped
// effect. While it stands, every ballot of every vote is put to the
// caster instead of the voter — the same ballot, in the same turn order
// and with the voter's extra votes (CR 701.38d), logged as the voter's
// vote — and the prompt says whose vote it is. A vote that begins later
// in the turn is covered, which is why this is a record and not a flag
// on one vote.
//
// Two of them: the newer record wins (CR 613.7), so the last caster
// chooses. A caster who leaves the game stops choosing and their
// ballots go back to the voters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3ad9287-ef13-4b66-9df8-b17e15bb4b0f",
		Name:         "Illusion of Choice",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ctx.Game.ChooseHowEachPlayerVotesThisTurnForEffect(ctx.Source(), ctx.Controller(),
				"Illusion of Choice — you choose how each player votes this turn")
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}

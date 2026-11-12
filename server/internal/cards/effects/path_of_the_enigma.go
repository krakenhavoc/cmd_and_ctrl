package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path of the Enigma — Sorcery {4}{U} (March of the Machine Commander,
// #5648):
//
//	"Target player draws four cards.
//	 Will of the Planeswalkers — Starting with you, each player votes for
//	 planeswalk or chaos. If planeswalk gets more votes, planeswalk. If
//	 chaos gets more votes or the vote is tied, chaos ensues."
//
// The spell has one target, so a target that has become illegal by the
// time it resolves means the whole spell does nothing, the vote included
// (CR 608.2b). Otherwise that player draws four, then the vote
// (WillOfThePlaneswalkers) waits for anything the draws ask (CR 608.2c).
// The vote happens at every table; planeswalking and chaos do something
// only at a table playing Planechase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "602f3203-91ae-423c-8a19-c6c1a5034f27",
		Name:         "Path of the Enigma",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 4}),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				item := ctx.Item
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return DrawCards{Player: item.Targets[0].ID, N: 4}.Apply(ctx)
			}),
			WillOfThePlaneswalkers{Card: "Path of the Enigma"},
		),
	})
}

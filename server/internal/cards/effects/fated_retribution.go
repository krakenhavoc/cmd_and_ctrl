package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fated Retribution — Instant {4}{W}{W}{W} (Born of the Gods,
// #5559):
//
//	"Destroy all creatures and planeswalkers. If it's your turn, scry 2."
//
// Slash the Ranks' sweep with no exemption, then the scry behind "if
// it's your turn" (IsYourTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28cfafdb-5014-44e3-9133-4554129c1aee",
		Name:         "Fated Retribution",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy},
			Scry:  2,
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DestroyAllMatching{Match: Or(Creature(), Planeswalker())}).Apply(ctx); err != nil {
				return err
			}
			if !IsYourTurn(ctx.Game, item.Controller) {
				return nil
			}
			return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}

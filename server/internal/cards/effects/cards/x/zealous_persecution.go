package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zealous Persecution — Instant {W}{B} (Alara Reborn, #5586):
//
//	"Until end of turn, creatures you control get +1/+1 and creatures your
//	 opponents control get -1/-1."
//
// Two until-end-of-turn layer 7c boosts over the creatures on the
// battlefield as it resolves (CR 611.2c): Inspiring Captain's +1/+1 to
// yours and Doomwake Giant's -1/-1 to your opponents'. The -1/-1 is a
// sweep of their X/1s.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccf8d919-1399-4bef-94ac-04c67f122828",
		Name:         "Zealous Persecution",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()),
				Power: 1, Toughness: 1,
				Label: "Zealous Persecution — +1/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OpponentControls()),
				Power: -1, Toughness: -1,
				Label: "Zealous Persecution — -1/-1",
			}.Apply(ctx)
		},
	})
}

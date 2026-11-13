package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reward the Faithful — Instant {W} (Scourge):
//
//	"Any number of target players each gain life equal to the greatest
//	 mana value among permanents you control."
//
// Any number of player targets, zero included (CR 601.2c), and one amount
// read as the spell resolves (scgB1GreatestManaValue) that each still-legal
// target gains.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fb7c848-377c-4476-9b3f-d95d145137f1",
		Name:         "Reward the Faithful",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("any number of target players").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := scgB1GreatestManaValue(ctx)
			if n <= 0 {
				return nil
			}
			for _, p := range kozilekTargetPlayers(ctx) {
				if err := (GainLife{Player: p, Amount: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gift of the Viper — Instant {G} (Modern Horizons 3, #5467):
//
//	"Put a +1/+1 counter, a reach counter, and a deathtouch counter on
//	 target creature. Untap it."
//
// Sagu Pummeler's counters with a deathtouch counter beside the reach
// one (CR 122.1b: each keyword counter gives the creature that keyword
// for as long as it stays), then Acrobatic Leap's untap. A target gone
// by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b1a59db-0ffc-490b-a881-c8192f3747db",
		Name:         "Gift of the Viper",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				for _, kind := range []string{game.CounterPlusOne, game.CounterReach, game.CounterDeathtouch} {
					if err := (AddCounter{Target: t.ID, Kind: kind, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return UntapTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}

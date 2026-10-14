package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Protection Magic — Instant {1}{W}:
//
//	"Put a shield counter on each of up to three target creatures. (If
//	 a creature with a shield counter would be dealt damage or
//	 destroyed, remove a shield counter from it instead.)"
//
// The counters do the rest through the CR 122.1c built-ins
// (game/shield_counters.go, #2000). Each target still legal on
// resolution gets one (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1380f0cc-e875-46ec-a802-c6684bf3965c",
		Name:         "Protection Magic",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterShield, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

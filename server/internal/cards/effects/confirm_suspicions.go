package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Confirm Suspicions — Instant {3}{U}{U} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Counter target spell.
//	 Investigate three times."
//
// Cancel's counter, then three Clues (CR 701.16a). The Clues are made
// even when the target was already gone: the spell is only countered if
// its target is still legal, but with an illegal target the whole spell
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a30d076b-c942-45e4-b943-3749ce955291",
		Name:         "Confirm Suspicions",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return Investigate{N: 3}.Apply(ctx)
		},
	})
}

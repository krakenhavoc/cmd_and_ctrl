package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Highspire Infusion — Instant {1}{G}:
//
//	"Target creature gets +3/+3 until end of turn. You get {E}{E} (two
//	 energy counters)."
//
// ADR 0129 PR 1. An illegal target means the spell does nothing at all,
// energy included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c86ed70b-ab09-402e-8aaf-6332b4749453",
		Name:         "Highspire Infusion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     3,
				Toughness: 3,
				Label:     "Highspire Infusion — +3/+3",
			}).Apply(ctx); err != nil {
				return err
			}
			return GetEnergy{N: 2}.Apply(ctx)
		},
	})
}

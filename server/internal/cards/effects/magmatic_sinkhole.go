package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magmatic Sinkhole — Instant {5}{R}:
//
//	"Delve. Magmatic Sinkhole deals 5 damage to target creature or planeswalker."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2da7f1b7-316b-4e8e-9fa3-5880b258d601",
		Name:         "Magmatic Sinkhole",
		Completeness: CompletenessFull,
		Delve:        true,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 5}.Apply(ctx)
		},
	})
}

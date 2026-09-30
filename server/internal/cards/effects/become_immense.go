package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Become Immense — Instant {5}{G}:
//
//	"Delve. Target creature gets +6/+6 until end of turn."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40e5b83e-5f53-4b15-8bfd-c0c0b8355a6f",
		Name:         "Become Immense",
		Completeness: CompletenessFull,
		Delve:        true,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     6,
				Toughness: 6,
				Label:     "Become Immense — +6/+6",
			}.Apply(ctx)
		},
	})
}

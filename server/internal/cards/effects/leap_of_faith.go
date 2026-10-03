package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leap of Faith — Instant {2}{W}:
//
//	"Target creature gains flying until end of turn. Prevent all damage
//	 that would be dealt to that creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the keyword grant, then the
// not-one-use shield pinned to the same creature.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ecf0b5ad-5989-4b99-9552-c5a24504f983",
		Name:         "Leap of Faith",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return thenShieldTheTarget(ctx, false, func(id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Leap of Faith — flying"}.Apply(ctx)
			})
		},
	})
}

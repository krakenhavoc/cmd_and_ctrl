package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Djeru's Resolve — Instant {W}:
//
//	"Untap target creature. Prevent all damage that would be dealt to it
//	 this turn.
//	 Cycling {2}"
//
// ADR 0108 §7, Delivery PR 7 (#1904): the untap, then the not-one-use
// shield pinned to the same creature for the rest of the turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c48ce123-bb07-4221-b3b3-0a458e101ca1",
		Name:         "Djeru's Resolve",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Activated:    []ActivatedAbility{Cycling("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return thenShieldTheTarget(ctx, false, func(id uuid.UUID) error {
				return UntapTarget{Target: id}.Apply(ctx)
			})
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Into the Fray — Instant — Arcane {R}:
//
//	"Target creature attacks this turn if able.
//	Splice onto Arcane {R}"
//
// The attack requirement Gogo, Mysterious Mime grants, until end of
// turn. Splice is ADR 0202 (#1989).
func init() {
	Register(Spec{
		OracleID:     "3d165dec-1821-4d3f-8407-8ce8bb347f8b",
		Name:         "Into the Fray",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{R}"),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return ScopedEffectFor{
				Target:   t.ID,
				Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Into the Fray — attacks this turn if able",
			}.Apply(ctx)
		},
	})
}

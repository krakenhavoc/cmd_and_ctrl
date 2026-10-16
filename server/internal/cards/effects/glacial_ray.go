package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glacial Ray — Instant — Arcane {1}{R}:
//
//	"Glacial Ray deals 2 damage to any target.
//	Splice onto Arcane {1}{R}"
//
// Spliced onto an Arcane spell, "Glacial Ray" is that spell (CR 702.47c):
// the damage is dealt by it, with its colour, so a creature with
// protection from red may be targeted from a blue spell (ADR 0202).
func init() {
	Register(Spec{
		OracleID:     "e65a8ab8-84a7-4fe2-b009-ceb87bec6b4e",
		Name:         "Glacial Ray",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{R}"),
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
		},
	})
}

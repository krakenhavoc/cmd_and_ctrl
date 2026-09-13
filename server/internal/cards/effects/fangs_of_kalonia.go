package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fangs of Kalonia — Sorcery {1}{G} (EDHREC rank 2517):
//
//	"Put a +1/+1 counter on target creature you control, then double
//	 the number of +1/+1 counters on each creature that had a +1/+1
//	 counter put on it this way.
//	 Overload {4}{G}{G} (You may cast this spell for its overload
//	 cost. If you do, change "target" in its text to "each.")"
//
// Two mana for +2/+2 on a bare creature, or "+1/+1 then double" on a
// counter-laden one; six mana and every creature you control gets
// it. The S22 overload machinery carries both halves — the price,
// and the deletion of the target clause, so the overloaded cast
// announces with no target and cannot be fizzled. The body is the
// same either way (b23GrowThenDoubleEach): first counters on every
// creature in the set, THEN the doubling, each placement through
// AddCounter so Doubling Season and Hardened Scales apply to both
// steps, as printed — a bare creature under Doubling Season ends
// with six.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d9bd4cc-564b-4fdf-9462-b4bb30583642",
		Name:         "Fangs of Kalonia",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{4}{G}{G}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return b23GrowThenDoubleEach(ctx, b23CreaturesYouControl(ctx.Game, ctx.Controller()))
			}
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return b23GrowThenDoubleEach(ctx, ids)
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theorix Charm — Instant {U}{B} (Reality Fracture, tracker #2795):
//
//	"Choose one —
//	 • Counter target noncreature spell unless its controller pays {2}.
//	 • Target creature gets -2/-2 until end of turn.
//	 • Mill three cards, then draw a card."
//
// The tax question goes to the countered spell's controller. The draw
// runs after the mill has landed (it is the mill's continuation).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf527bbd-e898-4aa9-909d-daec2f4b62ad",
		Name:         "Theorix Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Counter target noncreature spell unless its controller pays {2}.",
				TargetSpell("target noncreature spell", Noncreature()), rfSpellBCounterNoncreatureUnlessPays),
			ModeDoing("Target creature gets -2/-2 until end of turn.",
				TargetCreature("target creature"), BoostTheModesTarget(-2, -2, "Theorix Charm — -2/-2")),
			ModeDoing("Mill three cards, then draw a card.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return MillToZone{N: 3, Then: func(c *Context, _ []uuid.UUID) error {
						return DrawCards{Player: c.Controller(), N: 1}.Apply(c)
					}}.Apply(ctx)
				}),
		),
	})
}

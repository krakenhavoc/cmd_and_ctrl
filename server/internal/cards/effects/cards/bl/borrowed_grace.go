package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borrowed Grace — Instant {2}{W}:
//
//	"Escalate {1}{W} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or both —
//	 • Creatures you control get +2/+0 until end of turn.
//	 • Creatures you control get +0/+2 until end of turn."
//
// Escalate (CR 702.120a, #2126): {1}{W} for the second mode. Each
// bullet snapshots the creatures you control as it resolves (CR
// 611.2c), so a creature that arrives later gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e858d26f-db45-4c56-b40a-bbaeedbf5549",
		Name:         "Borrowed Grace",
		Completeness: effects.CompletenessFull,
		Modes: effects.Escalating(effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Creatures you control get +2/+0 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "Borrowed Grace"}.Apply(ctx)
				}),
			effects.ModeDoing("Creatures you control get +0/+2 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Toughness: 2, Label: "Borrowed Grace"}.Apply(ctx)
				}),
		), effects.EscalateMana("{1}{W}")),
	})
}

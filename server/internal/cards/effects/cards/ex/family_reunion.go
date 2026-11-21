package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Family Reunion — Instant {1}{W} (The Lost Caverns of Ixalan, #4470):
//
//	"Choose one —
//	 • Creatures you control get +1/+1 until end of turn.
//	 • Creatures you control gain hexproof until end of turn."
//
// The mode is chosen as the spell is cast (CR 700.2a). Either bullet
// changes characteristics, so the creatures it affects are the ones its
// controller controls as it resolves, and a creature that enters later
// gets nothing (CR 611.2c). Hexproof is CR 702.11.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8972de18-2c49-4162-a285-264a44c0e675",
		Name:         "Family Reunion",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Creatures you control get +1/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
						Label: "Family Reunion — creatures you control get +1/+1",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
			effects.ModeWithPurpose(effects.ModeDoing("Creatures you control gain hexproof until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"hexproof"},
						Label:    "Family Reunion — creatures you control gain hexproof",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerProtect}),
		),
	})
}

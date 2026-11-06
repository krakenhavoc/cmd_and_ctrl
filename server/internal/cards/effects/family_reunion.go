package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8972de18-2c49-4162-a285-264a44c0e675",
		Name:         "Family Reunion",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Creatures you control get +1/+1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
						Label: "Family Reunion — creatures you control get +1/+1",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
			ModeWithPurpose(ModeDoing("Creatures you control gain hexproof until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"hexproof"},
						Label:    "Family Reunion — creatures you control gain hexproof",
					}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerProtect}),
		),
	})
}

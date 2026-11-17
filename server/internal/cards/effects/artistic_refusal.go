package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artistic Refusal — Instant {4}{U}{U} (March of the Machine):
//
//	"Convoke
//	 Choose one or both —
//	 • Counter target spell.
//	 • Draw two cards, then discard a card."
//
// Convoke is Spec.TapCost. The two bullets are one "choose one or both"
// (CR 700.2) in printed order; a spell that has left the stack is skipped
// (CR 608.2b), so the draw and discard still happen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff192482-3808-4291-a31e-3dfaeb49cda2",
		Name:         "Artistic Refusal",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Counter target spell.", TargetSpell("target spell"), CounterTheModesTarget),
			ModeWithPurpose(ModeDoing("Draw two cards, then discard a card.", nil, drawTwoThenDiscardOne),
				game.Purpose{Draws: 2, Discards: 1}),
		),
	})
}

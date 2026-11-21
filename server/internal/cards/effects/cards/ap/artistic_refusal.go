package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ff192482-3808-4291-a31e-3dfaeb49cda2",
		Name:         "Artistic Refusal",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Counter target spell.", effects.TargetSpell("target spell"), effects.CounterTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Draw two cards, then discard a card.", nil, effects.DrawTwoThenDiscardOne),
				game.Purpose{Draws: 2, Discards: 1}),
		),
	})
}

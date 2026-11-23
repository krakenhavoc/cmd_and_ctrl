package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Charm — Instant {B}{G}{U}:
//
//	"Choose one —
//	 • Destroy target monocolored creature.
//	 • Destroy target artifact or enchantment.
//	 • Draw two cards, then discard a card."
//
// "Monocolored" is exactly one colour as the creature is NOW (layer 5
// applies), so a colourless creature and a multicoloured one are both
// illegal targets — the Vanishing Verse reading (b41Monocolored).
//
// Mode 2 draws first and queues the discard afterwards, which matters:
// the discard prompt is built from the post-draw hand, so a card just
// drawn is a legal discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46ed38d1-e642-4cea-99ed-a9c17fd982b1",
		Name:         "Sultai Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target monocolored creature.",
				effects.TargetCreature("target monocolored creature", effects.B41Monocolored()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Draw two cards, then discard a card.", nil, effects.DrawTwoThenDiscardOne),
		),
	})
}

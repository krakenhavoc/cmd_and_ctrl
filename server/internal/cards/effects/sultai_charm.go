package effects

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
	Register(Spec{
		OracleID:     "46ed38d1-e642-4cea-99ed-a9c17fd982b1",
		Name:         "Sultai Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target monocolored creature.",
				TargetCreature("target monocolored creature", b41Monocolored()),
				DestroyTheModesTarget),
			ModeDoing("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				DestroyTheModesTarget),
			ModeDoing("Draw two cards, then discard a card.", nil, drawTwoThenDiscardOne),
		),
	})
}

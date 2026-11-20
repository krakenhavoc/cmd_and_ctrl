package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witherbloom Apprentice — Creature — Human Druid {B}{G}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 each opponent loses 1 life and you gain 1 life."
//
// Professor Onyx's magecraft drain at 1: each opponent loses 1 and you
// gain a flat 1, not the total lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "696f554d-0485-48a5-9273-3f6fb7d16a5d",
		Name:         "Witherbloom Apprentice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Witherbloom Apprentice — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}

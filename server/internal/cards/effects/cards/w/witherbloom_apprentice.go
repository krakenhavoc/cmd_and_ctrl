package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "696f554d-0485-48a5-9273-3f6fb7d16a5d",
		Name:         "Witherbloom Apprentice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Witherbloom Apprentice — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}

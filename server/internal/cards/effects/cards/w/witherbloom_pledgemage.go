package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witherbloom Pledgemage — Creature — Treefolk Warlock {3}{B/G}{B/G}, 5/5
// (Strixhaven: School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell, you
//	 gain 1 life."
//
// The shared Magecraft trigger over a plain life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f10039ae-2695-4b0a-82b2-82ee79a992c7",
		Name:         "Witherbloom Pledgemage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Witherbloom Pledgemage — you gain 1 life (magecraft)", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}

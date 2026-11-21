package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Well of Life — Artifact {4} (Prophecy, #5794):
//
//	"At the beginning of your end step, if you control no untapped lands,
//	 you gain 2 life."
//
// An intervening "if" (CR 603.4, no_untapped_lands.go): checked as its
// controller's end step begins and again as the trigger resolves, so a
// land untapped in response means no life is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e05d1a2-ad53-4f2e-aaea-7ec464b07c04",
		Name:         "Well of Life",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfNoUntappedLands("Well of Life — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}

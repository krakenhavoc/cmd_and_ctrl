package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shopkeeper's Bane — Creature — Badger Pest {2}{G}, 4/2 (Secrets of
// Strixhaven, #4456):
//
//	"Trample
//	 Whenever this creature attacks, you gain 2 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e7572317-6bec-4fb0-81ac-76aec04f14ad",
		Name:            "Shopkeeper's Bane",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Shopkeeper's Bane — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}

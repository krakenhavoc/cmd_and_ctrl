package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shopkeeper's Bane — Creature — Badger Pest {2}{G}, 4/2 (Secrets of
// Strixhaven, #4456):
//
//	"Trample
//	 Whenever this creature attacks, you gain 2 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7572317-6bec-4fb0-81ac-76aec04f14ad",
		Name:            "Shopkeeper's Bane",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Shopkeeper's Bane — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}

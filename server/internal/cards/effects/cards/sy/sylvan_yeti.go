package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvan Yeti — {2}{G}{G} Creature — Yeti */4 (Starter 1999, S82 OLD batch 702, #5620):
//
//	"Sylvan Yeti's power is equal to the number of cards in your hand."
//
// A characteristic-defining ability (CR 604.3, 613.4a) read live, as Syr
// Elenora's (PowerIsCardsInYourHand).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6493960b-4bbb-4ef0-8efb-abb122b1697d",
		Name:         "Sylvan Yeti",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.PowerIsCardsInYourHand(), "Sylvan Yeti's power is equal to the number of cards in your hand."),
		},
	})
}

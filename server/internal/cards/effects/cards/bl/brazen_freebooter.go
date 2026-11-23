package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brazen Freebooter — Creature — Human Pirate {3}{R}, 3/3 (Rivals of
// Ixalan, #5531):
//
//	"When this creature enters, create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// Redcap Thief's entry: the Treasure is the catalog's printed token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13e8d063-eece-4f38-83f6-02403130defa",
		Name:         "Brazen Freebooter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Brazen Freebooter — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}

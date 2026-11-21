package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Emissary — Creature — Zombie Warrior {1}{B}, 1/1 (Fate Reforged,
// #5552):
//
//	"When this creature dies, manifest the top card of your library. (Put
//	 that card onto the battlefield face down as a 2/2 creature. Turn it
//	 face up any time for its mana cost if it's a creature card.)"
//
// Whisperwood Elemental's granted dies trigger, printed: the top card of
// your library is put onto the battlefield face down as a 2/2 (CR 701.40a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0274f286-46a1-4028-878a-88c2146a895c",
		Name:         "Sultai Emissary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Sultai Emissary — manifest the top card of your library", effects.Do(effects.Manifest{N: 1})),
		},
	})
}

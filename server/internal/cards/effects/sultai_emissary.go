package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0274f286-46a1-4028-878a-88c2146a895c",
		Name:         "Sultai Emissary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Sultai Emissary — manifest the top card of your library", Do(Manifest{N: 1})),
		},
	})
}

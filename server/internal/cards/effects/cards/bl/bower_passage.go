package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bower Passage — Enchantment {1}{G} (Avacyn Restored, S82 OLD, #5570):
//
//	"Creatures with flying can't block creatures you control."
//
// A block rule carried by the enchantment (CR 509.1b): a creature with flying,
// read through the layered view, may be declared as the blocker of no attacker
// attacker controlled by this
// enchantment's controller. It binds every player's fliers, yours too, and
// ends when the enchantment leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f91ef713-5289-4901-9291-0a5afe6f997b",
		Name:         "Bower Passage",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnMatching(effects.HasKeyword("flying")), effects.ControlledBySourceController(),
				"creatures with flying can't block creatures the Bower Passage's controller controls"),
		},
	})
}

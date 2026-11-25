package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul-Strike Technique — Enchantment — Aura {1}{W} (Modern Horizons):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has vigilance.
//	 When enchanted creature dies, manifest the top card of your library.
//	 (Put it onto the battlefield face down as a 2/2 creature. Turn it face
//	 up any time for its mana cost if it's a creature card.)"
//
// The pump is PumpAttached and vigilance GrantToAttached, both read through
// the layers. The dies trigger is Minion's Return's (WhenEnchantedCreatureDies,
// harvested while the Aura is still attached), and the Aura's controller, not
// the creature's, manifests the top card of their own library (CR 701.40a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a6b32cd-a0f5-4e3e-9069-b525d485f84f",
		Name:         "Soul-Strike Technique",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1), effects.GrantToAttached("vigilance")},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Soul-Strike Technique — manifest the top card of your library", effects.Do(effects.Manifest{N: 1})),
		},
	})
}

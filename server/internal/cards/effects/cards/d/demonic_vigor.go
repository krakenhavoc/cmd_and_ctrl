package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Vigor — Enchantment — Aura {B} (Dominaria, S82 OLD, #5530):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 When enchanted creature dies, return that card to its owner's hand."
//
// A layer 7c +1/+1 on the host, and a dies trigger on the host's death
// (Minion's Return's WhenEnchantedCreatureDies) that brings the card back from
// the graveyard it just reached. A token has no card to return, and a card
// that has left the graveyard by resolution stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00cb93a8-5d33-4c6a-9e2a-4b3560f76f3f",
		Name:         "Demonic Vigor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Demonic Vigor — return that card to its owner's hand", effects.S82OldB212ReturnDiedCardToOwnersHand),
		},
	})
}

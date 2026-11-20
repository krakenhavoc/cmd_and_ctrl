package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moldervine Cloak — Enchantment — Aura {2}{G} (Ravnica: City of Guilds,
// #5600):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 Dredge 2 (If you would draw a card, you may mill two cards instead.
//	 If you do, return this card from your graveyard to your hand.)"
//
// The +3/+3 is PumpAttached on the host (layer 7c), and dredge is the
// engine's draw replacement while the Aura is in its owner's graveyard (CR
// 702.52).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd106f2d-5ede-43f0-928a-1ab392b90854",
		Name:         "Moldervine Cloak",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Replacements: []game.ReplacementEffect{Dredge(2)},
	})
}

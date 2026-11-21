package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taste for Mayhem — Enchantment — Aura {R} (Ravnica Remastered):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0.
//	 Hellbent — Enchanted creature gets an additional +2/+0 as long as you
//	 have no cards in hand."
//
// Two layer 7c modifies; the second also asks the Aura's controller's hand
// (its size is public), re-read every time the layers are computed.
//
// No simplification.
func init() {
	hellbentPump := game.StaticAbility{
		Label:    "Hellbent — Attached creature gets an additional +2/+0 as long as you have no cards in hand.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !AttachedToSource(target, g, source) {
				return false
			}
			p := g.PlayerByIDForEffect(source.Controller)
			return p != nil && (p.Hand == nil || p.Hand.Size() == 0)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += 2
		},
	}
	Register(Spec{
		OracleID:     "60e7517f-7cc5-4073-9f2c-3e6acb891128",
		Name:         "Taste for Mayhem",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(2, 0), hellbentPump},
	})
}

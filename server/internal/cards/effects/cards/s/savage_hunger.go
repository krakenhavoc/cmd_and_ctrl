package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Savage Hunger — Enchantment — Aura {2}{G} (Shards of Alara, S82 OLD batch
// 517, #5588):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and has trample.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Aether Tunnel's +1/+0 on the host with Angelic Gift's granted keyword; both
// ride the attachment, so they end when the Aura does. Cycling is the
// keyword's own hand ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e608f48-d3a6-422c-839d-f06bb8a32014",
		Name:         "Savage Hunger",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}

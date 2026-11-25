package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghostly Possession — Enchantment — Aura {2}{W} (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 Prevent all combat damage that would be dealt to and dealt by enchanted
//	 creature."
//
// Air Bladder's flying grant, and Fog Bank's standing prevention (CR 615,
// ev.IsCombatDamage only) pointed at whatever the Aura enchants as the
// damage would be dealt: combat damage whose recipient or whose source is
// that creature is prevented. Noncombat damage is not touched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8104a51d-f07b-45d6-aaa7-c6386d6de8bf",
		Name:         "Ghostly Possession",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || !ev.IsCombatDamage {
					return false
				}
				host := g.AttachedHostOf(src)
				return host != nil && (ev.DamageTarget == host.InstanceID || ev.DamageSource == host.InstanceID)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Ghostly Possession — prevent all combat damage dealt to and by enchanted creature",
		}},
	})
}

package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armament of Nyx — Enchantment — Aura {2}{W} (Journey into Nyx):
//
//	"Enchant creature
//	 Enchanted creature has double strike as long as it's an enchantment.
//	 Otherwise, prevent all damage that would be dealt by enchanted creature."
//
// Both halves read the host's type through the layered view each time: the
// grant is a layer 6 static under the condition, and the prevention (CR 615.1)
// applies only to damage dealt by a host that is not an enchantment.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a4d16eb-343e-41ca-9363-90dc79da4769",
		Name:         "Armament of Nyx",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.B43GrantToAttachedWhen(func(host *game.Card, g *game.Game, source *game.Card) bool {
				return effects.AttachedToSource(host, g, source) && host.IsEnchantment()
			}, "double strike"),
		},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 ||
					src.AttachedTo.ID == uuid.Nil || ev.DamageSource != src.AttachedTo.ID {
					return false
				}
				host, ok := g.LookupCardForEffect(src.AttachedTo.ID)
				return ok && !host.IsEnchantment()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Armament of Nyx — prevent all damage dealt by enchanted creature unless it's an enchantment",
		}},
	})
}

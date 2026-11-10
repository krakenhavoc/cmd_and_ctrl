package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cryoshatter — Enchantment — Aura {U} (Edge of Eternities, #4461):
//
//	"Enchant creature
//	 Enchanted creature gets -5/-0.
//	 When enchanted creature becomes tapped or is dealt damage, destroy it."
//
// The -5/-0 is a layer 7c change on the host (pumpEnchantedCreature). The
// trigger is one printed ability on three events: a tap (the tap event, and
// the attack that taps it, EnchantedPermanentBecameTapped) and damage dealt
// to the host by any source. Simultaneous damage from several sources is
// one batch and one trigger. "It" is the creature the event names, so the
// destroy still lands if the Aura has left by then; a creature that is
// already gone is not destroyed again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "865715e9-6759-4e01-803d-702c959e386e",
		Name:         "Cryoshatter",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{pumpEnchantedCreature(-5, 0)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack, game.EventDealDamage},
				func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventDealDamage {
						return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
					}
					return EnchantedPermanentBecameTapped(ev, source, ch, g)
				},
				"Cryoshatter — destroy enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					id := ev.CardID
					if ev.Kind == game.EventDealDamage {
						id = ev.Target
					}
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
						return nil
					}
					return DestroyTarget{Target: id}.Apply(NewContext(g, item))
				})),
		},
	})
}

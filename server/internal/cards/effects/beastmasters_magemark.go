package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beastmaster's Magemark — Enchantment — Aura {2}{G} (Guildpact, #5599):
//
//	"Enchant creature
//	 Creatures you control that are enchanted get +1/+1.
//	 Whenever a creature you control that's enchanted becomes blocked, it
//	 gets +1/+1 until end of turn for each creature blocking it."
//
// The anthem is A Tale for the Ages' at +1/+1: a layer 7c modify on every
// creature the Aura's controller controls that has an Aura attached,
// whoever controls that Aura, re-read each layer pass. The creature it
// enchants is one of them, so two Magemarks on one creature give it
// +2/+2 each, and a Magemark on an opponent's creature pumps only your
// own enchanted creatures.
//
// The trigger watches each attacker that becomes blocked (CR 509.1h),
// once per attacker however many blockers, and asks whether it is a
// creature you control with an Aura on it then. The count is the
// creatures blocking it as the trigger resolves (blockersOf, Mystic Doom
// Sandwich's read).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6889f57-2965-4a82-8556-fd91c6227058",
		Name:         "Beastmaster's Magemark",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{{
			Label:    "Creatures you control that are enchanted get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && isEnchantedByAura(g, target.InstanceID)
			},
			Apply: pumpBy(1, 1),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesBlocked, yourEnchantedCreatureBecameBlocked,
				"Beastmaster's Magemark — it gets +1/+1 until end of turn for each creature blocking it",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					attacker := item.Trigger.Event.Target
					n := len(blockersOf(g, attacker))
					if n <= 0 {
						return nil
					}
					return BoostUntilEOT{Target: attacker, Power: n, Toughness: n, Label: "Beastmaster's Magemark"}.Apply(NewContext(g, item))
				}),
		},
	})
}

// yourEnchantedCreatureBecameBlocked is "whenever a creature you control
// that's enchanted becomes blocked": the attacker the event names is a
// creature the source's controller controls with an Aura attached.
func yourEnchantedCreatureBecameBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsCreature() && c.Controller == source.Controller && isEnchantedByAura(g, c.InstanceID)
}

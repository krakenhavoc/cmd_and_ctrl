package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ordeal of Purphoros — Enchantment — Aura {1}{R} (Jumpstart 2022):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, put a +1/+1 counter on it. Then if it
//	 has three or more +1/+1 counters on it, sacrifice this Aura.
//	 When you sacrifice this Aura, it deals 3 damage to any target."
//
// Ordeal of Nylea's two abilities. The attack trigger puts the counter and
// then, in the same resolution, counts the creature's +1/+1 counters as they
// are now. The sacrifice trigger watches EventSacrifice, which fires while the
// Aura is still on the battlefield, for THIS Aura sacrificed by its
// controller, so it triggers for any sacrifice, not only its own; an Aura that
// falls off as a state-based action is not sacrificed. The target is chosen
// as the trigger goes on the stack and the Aura is the damage source as it
// last existed (CR 608.2h).
//
// No simplification.
func init() {
	const name = "Ordeal of Purphoros"
	effects.Register(effects.Spec{
		OracleID:     "652d282d-6dec-4043-a531-480f12ba471e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks(name+" — put a +1/+1 counter on it, then sacrifice this Aura if it has three or more",
				effects.OrdealAttackCounterThenSacrifice),
			effects.Targeting(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventSacrifice},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID && ev.Actor == source.Controller
				},
				Key:    name + " — it deals 3 damage to any target",
				Effect: effects.DamageToFirstTarget(3),
			}, effects.TargetAny()),
		},
	})
}

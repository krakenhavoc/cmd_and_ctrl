package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kosei, Penitent Warlord — Legendary Creature — Ogre Samurai {1}{G}{G}, 0/5
// (Kamigawa: Neon Dynasty):
//
//	"As long as Kosei is enchanted, equipped, and has a counter on it, Kosei
//	 has "Whenever Kosei deals combat damage to an opponent, you draw that
//	 many cards and Kosei deals that much damage to each other opponent.""
//
// A layer-6 grant of a trigger bundle to Kosei alone (Agent of the Iron
// Throne's shape), gated on all three conditions at once: an Aura of any
// controller attached ("enchanted" does not say "by an Aura you control"),
// an Equipment attached, and a counter of any kind. The gate reads attachments
// and counters only, so it is safe inside a layer pass, and the grant drops
// the moment any of the three goes. The trigger is Amarant Coral's: how much
// landed, and on whom, are facts of the damage event; the draw comes first,
// then the damage to each other opponent, as one damage instance.
//
// No simplification.
const koseiGrant = "kosei-penitent-warlord/combat-damage-draw"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "41f7e725-bdb8-4c3a-adc7-a8184ec7c394",
		Name:         "Kosei, Penitent Warlord",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: koseiGrant,
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B38ThisDealtCombatDamageToAnOpponent(ev, source, g)
				},
				Key:    "Kosei, Penitent Warlord — you draw that many cards and Kosei deals that much damage to each other opponent",
				Effect: koseiCombatDamage,
			}},
			Text: "Whenever Kosei deals combat damage to an opponent, you draw that many cards and Kosei deals that much damage to each other opponent.",
		}},
		Static: []game.StaticAbility{func() game.StaticAbility {
			s := effects.GrantAbilities(koseiEnchantedEquippedAndCountered, koseiGrant)
			s.Label = "As long as Kosei is enchanted, equipped, and has a counter on it, Kosei has \"Whenever Kosei deals combat damage to an opponent, you draw that many cards and Kosei deals that much damage to each other opponent.\""
			return s
		}()},
	})
}

// koseiEnchantedEquippedAndCountered is the gate: the target is Kosei, with
// some Aura attached, some Equipment attached and a counter on it.
func koseiEnchantedEquippedAndCountered(target *game.Card, g *game.Game, source *game.Card) bool {
	if target.InstanceID != source.InstanceID || g == nil || g.Battlefield == nil {
		return false
	}
	counter := false
	for _, n := range target.Counters {
		if n > 0 {
			counter = true
		}
	}
	if !counter {
		return false
	}
	enchanted, equipped := false, false
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if !a.IsAttachedTo(target.InstanceID) {
			continue
		}
		if a.HasSubtype("Equipment") {
			equipped = true
		}
		if a.IsAura() {
			enchanted = true
		}
	}
	return enchanted && equipped
}

// koseiCombatDamage draws that many cards, then deals that much damage to
// each other opponent.
func koseiCombatDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	if err := (effects.DrawCards{Player: item.Controller, N: item.Trigger.Event.Amount}).Apply(effects.NewContext(g, item)); err != nil {
		return err
	}
	return effects.ThatMuchToEachOpponentButTheOneHit(g, item)
}

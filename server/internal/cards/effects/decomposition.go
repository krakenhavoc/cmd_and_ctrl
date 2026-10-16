package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decomposition — Enchantment — Aura {1}{G}:
//
//	"Enchant black creature
//	 Enchanted creature has 'Cumulative upkeep—Pay 1 life.'
//	 When enchanted creature dies, its controller loses 2 life."
//
// #3869. Mana Chains with a life payment: a layer-6 grant of the
// CumulativeUpkeepGrantPaying bundle with LifePayment(1), so at N age
// counters the creature's controller is asked to pay N life, through the
// life-cost gate (CR 119.4: a player with less life than that can't pay,
// and the creature is sacrificed). The dies trigger is the Aura's, and
// "its controller" is the controller the creature had as it died (CR
// 603.10a), read from its last-known record.
//
// "black" is read through the enchant walk's word lens (OfColor).
//
// No simplification.
const decompositionGrant = "decomposition/cumulative-upkeep-pay-1-life"

func init() {
	Register(Spec{
		OracleID:     "62817ef5-9b8d-4bca-bdd4-ab3bc4e063d0",
		Name:         "Decomposition",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(OfColor("B")),
		Grants: []AbilityGrant{
			CumulativeUpkeepGrantPaying(decompositionGrant, LifePayment(1), "Cumulative upkeep—Pay 1 life"),
		},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesToAttached(decompositionGrant), "Enchanted creature has \"Cumulative upkeep—Pay 1 life.\""),
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Decomposition — its controller loses 2 life", decompositionDrain),
		},
	})
}

// decompositionDrain is "its controller loses 2 life": whoever controlled
// the creature as it died.
func decompositionDrain(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	dead := item.Trigger.Event.CardID
	info, ok := g.LastKnownPermanentForEffect(dead)
	if !ok {
		return nil
	}
	if p := g.PlayerByIDForEffect(info.Controller); p == nil || p.Eliminated {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, info.Controller, -2)
}

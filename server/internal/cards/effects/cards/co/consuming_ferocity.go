package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consuming Ferocity — Enchantment — Aura {1}{R} (Mirage, S82 OLD batch 811,
// #5633):
//
//	"Enchant non-Wall creature
//	 Enchanted creature gets +1/+0.
//	 At the beginning of your upkeep, put a +1/+0 counter on enchanted
//	 creature. If that creature has three or more +1/+0 counters on it, it
//	 deals damage equal to its power to its controller, then destroy that
//	 creature and it can't be regenerated."
//
// The counters are the creature's own (CR 122.1a: a +1/+0 counter changes its
// power), so they stay and count when the Aura is replaced. One printed
// trigger: counter, then the check, the damage and the destruction in that
// order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8efbe68c-622c-4986-a375-7e0bf780eb2e",
		Name:         "Consuming Ferocity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.Not(effects.OfCreatureType("Wall"))),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Consuming Ferocity — put a +1/+0 counter on enchanted creature", effects.S82OldB811ConsumingFerocityUpkeep),
		},
	})
}

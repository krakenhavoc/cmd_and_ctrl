package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niv-Mizzet, Dracogenius — Legendary Creature — Dragon Wizard {2}{U}{U}{R}{R},
// 5/5 (Commander 2017, S82 older sets):
//
//	"Flying
//	 Whenever Niv-Mizzet deals damage to a player, you may draw a card.
//	 {U}{R}: Niv-Mizzet deals 1 damage to any target."
//
// The trigger is Abyssal Specter's (any damage, combat or not, to a player),
// with the "may draw" of Aven Fisher. The ping is an ordinary targeted
// activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "899d58dc-60b2-4013-9f91-c1888c585d66",
		Name:            "Niv-Mizzet, Dracogenius",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventDealDamage, effects.ThisDealtDamageToAPlayer,
				"Niv-Mizzet, Dracogenius — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Niv-Mizzet, Dracogenius — draw a card?", 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{R}: Niv-Mizzet deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.ManaCost("{U}{R}"),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}

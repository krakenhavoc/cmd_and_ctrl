package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruric Thar, the Unbowed — Legendary Creature — Ogre Warrior {4}{R}{G},
// 6/6 (Commander Legends):
//
//	"Reach, vigilance
//	 Ruric Thar attacks each combat if able.
//	 Whenever a player casts a noncreature spell, Ruric Thar deals 6 damage
//	 to that player."
//
// Reach and vigilance are printed keywords; the attack requirement is
// AttacksEachCombat (Barricade Breaker's). The trigger is Kambal's cast
// watcher over any player's noncreature spell, its controller's included
// (a land is not cast, so it never triggers), and the damage goes to the
// player who cast it (Eidolon of the Great Revel's), as Ruric Thar last
// existed (CR 608.2h) if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ed13a89-4284-4c01-8099-75e2d68c8c63",
		Name:            "Ruric Thar, the Unbowed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source == nil {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			}, "Ruric Thar — 6 damage to that player", effects.ScgB1DamageThatPlayer(6)),
		},
	})
}

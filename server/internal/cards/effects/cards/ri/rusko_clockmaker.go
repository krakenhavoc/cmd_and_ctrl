package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rusko, Clockmaker — Legendary Creature — Human Artificer {2}{U}{B},
// 3/3 (Mystery Booster 2, #5948):
//
//	"When Rusko, Clockmaker enters the battlefield, conjure a card named
//	 Midnight Clock onto the battlefield.
//	 Whenever you cast a noncreature spell, put an hour counter on each
//	 permanent you control named Midnight Clock. Each opponent loses 1
//	 life and you gain 1 life."
//
// The conjured Midnight Clock is a real card (ADR 0099 amendment
// 2026-11-24), so its own "when the twelfth hour counter is put on" sees
// the counters this puts on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccb456a0-3abd-4bde-a453-20d2da210e05",
		Name:         "Rusko, Clockmaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rusko, Clockmaker — conjure a card named Midnight Clock onto the battlefield",
				effects.Do(effects.CheckConjure("Rusko, Clockmaker", effects.Conjure{Names: []string{"Midnight Clock"}, Zone: game.ZoneBattlefield}))),
			effects.On(game.EventCast, effects.YouCast(effects.Noncreature()),
				"Rusko, Clockmaker — an hour counter on each Midnight Clock you control; each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.CounterOnEachPermanentYouControlNamed{Name: "Midnight Clock", Kind: "hour"},
					effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}

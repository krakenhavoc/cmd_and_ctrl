package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Brushstroke — Enchantment {1}{B}{B} (Mystery Booster 2,
// #5948):
//
//	"When Sanguine Brushstroke enters the battlefield, create a Blood
//	 token and conjure a card named Blood Artist onto the battlefield.
//	 Whenever you sacrifice a Blood token, each opponent loses 1 life and
//	 you gain 1 life."
//
// The Blood token comes first, then the conjured Blood Artist, a real
// card (ADR 0099 amendment 2026-11-24) that enters through the ordinary
// entry. The sacrifice trigger is Gluttonous Guest's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2c6502d-fefd-4c9f-9d7b-6dfcc43316e1",
		Name:         "Sanguine Brushstroke",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sanguine Brushstroke — create a Blood token and conjure a Blood Artist onto the battlefield",
				effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1},
					effects.CheckConjure("Sanguine Brushstroke", effects.Conjure{Names: []string{"Blood Artist"}, Zone: game.ZoneBattlefield}))),
			effects.On(game.EventSacrifice, effects.VowB04YouSacrificedBlood, "Sanguine Brushstroke — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}

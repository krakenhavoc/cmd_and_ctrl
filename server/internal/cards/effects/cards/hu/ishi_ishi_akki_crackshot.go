package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ishi-Ishi, Akki Crackshot — Legendary Creature — Goblin Warrior {1}{R},
// 1/1 (Betrayers of Kamigawa, #5602):
//
//	"Whenever an opponent casts a Spirit or Arcane spell, Ishi-Ishi deals 2
//	 damage to that player."
//
// Mindsparker's cast trigger with the ki cycle's Spirit-or-Arcane spell
// test. It triggers on the cast, so it resolves first and happens even if
// the spell is countered. "That player" is the caster, read off the
// item's carried event; an Ishi-Ishi that left in response still deals the
// damage, named as the departed object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30265fce-28e6-4994-83f7-0e4c4b22df35",
		Name:         "Ishi-Ishi, Akki Crackshot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.ChkB1SpiritOrArcane()),
				"Ishi-Ishi, Akki Crackshot — 2 damage to that player",
				effects.SourceDealsTwoToThatPlayer),
		},
	})
}

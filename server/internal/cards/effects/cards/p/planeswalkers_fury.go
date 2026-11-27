package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planeswalker's Fury — Enchantment {2}{R} (Planeshift, S82 OLD #5615):
//
//	"{3}{R}: Target opponent reveals a card at random from their hand. This
//	 enchantment deals damage equal to that card's mana value to that player.
//	 Activate only as a sorcery."
//
// Friendly Fire's random reveal (the game's keyed random stream, shown to the
// table), then damage from the enchantment, as many as the revealed card's mana
// value. An empty hand reveals nothing and deals no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e482522-2eef-4f6c-887a-fcc51be61dc2",
		Name:         "Planeswalker's Fury",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{R}: Target opponent reveals a card at random from their hand. This enchantment deals damage equal to that card's mana value to that player. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{R}"),
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.Instruction(effects.OldB640RevealThenDamage)),
		}},
	})
}

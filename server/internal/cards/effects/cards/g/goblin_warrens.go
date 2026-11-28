package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Warrens — Enchantment {2}{R} (Fallen Empires, S82 OLD batch 852, #5637):
//
//	"{2}{R}, Sacrifice two Goblins: Create three 1/1 red Goblin creature tokens."
//
// Two Goblins are sacrificed as the ability is activated (Blighted Shaman's
// sacrifice cost over a subtype, any two Goblins you control, tokens and the
// Warrens' own new tokens included); the three tokens are made as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1161505-1623-4475-a88e-2e48072cc2f1",
		Name:         "Goblin Warrens",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice two Goblins: Create three 1/1 red Goblin creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeN(2, "two Goblins", effects.HasSubtype("Goblin"))),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin"), N: 3}),
		}},
	})
}

package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Furrier — Creature — Goblin Warrior {1}{R}, 2/2 (Coldsnap, #5597):
//
//	"Prevent all damage that this creature would deal to snow creatures."
//
// A standing CR 615 prevention of this creature's damage, combat or not, whose
// recipient is a snow creature on the battlefield as the damage would be dealt.
// Damage to a player, a planeswalker or a creature that is not snow is not
// touched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f086b72-6e7c-40d4-b995-d5f94c494462",
		Name:         "Goblin Furrier",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.S82OldB567PreventDamageDealtBy(
			"Goblin Furrier — prevent all damage it would deal to snow creatures",
			func(recipient game.Card) bool { return recipient.IsCreature() && recipient.HasSupertype("snow") })},
	})
}

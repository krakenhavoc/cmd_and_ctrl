package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirited Companion — Enchantment Creature — Dog, {1}{W}, 1/1
// (EDHREC rank 913):
//
//	"When this creature enters, draw a card."
//
// A cantrip on a body that is also an enchantment, which is why an
// enchantress deck plays it: it draws once on cast (Sythis, Mesa
// Enchantress) and once on entry, and it turns on constellation.
// Mulldrifter's trigger with a one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c5f0d91-9d86-4e66-94fd-4af93ad01838",
		Name:         "Spirited Companion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spirited Companion — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

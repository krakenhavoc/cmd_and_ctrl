package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metathran Elite — Creature — Metathran Soldier {1}{U}{U}, 2/3 (Urza's Destiny,
// S82 #5621):
//
//	"This creature can't be blocked as long as it's enchanted."
//
// A block rule on the Elite alone, read when the block is checked: it is
// enchanted while an Aura is attached to it (CR 303.4), so the rule follows the
// Aura on and off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8a1b3dd-bc36-4d97-8f93-1151446f04e2",
		Name:         "Metathran Elite",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return effects.IsEnchantedByAura(g, c.InstanceID)
		})},
	})
}

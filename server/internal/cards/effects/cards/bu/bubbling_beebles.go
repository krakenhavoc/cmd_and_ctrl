package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bubbling Beebles — Creature — Beeble {4}{U}, 3/3 (Urza's Destiny, S82 #5621):
//
//	"This creature can't be blocked as long as defending player controls an
//	 enchantment."
//
// A block rule on the Beebles alone (Waterlogged Hulk's shape) whose condition
// is read when the block is checked: the player the attacker is attacking
// controls an enchantment (post-layer types). Planeswalker and battle attacks
// read the defending player the same way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75903c3b-2dbe-4662-b6a8-39599e49adde",
		Name:         "Bubbling Beebles",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			defender := g.DefendingPlayerForAttackerForEffect(c.InstanceID)
			return defender != uuid.Nil && effects.B40ControlsA(g, defender, effects.Enchantment())
		})},
	})
}

package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Insight — Enchantment {2}{U} (Tempest, S82 OLD batch 766, #5628):
//
//	"Whenever an opponent casts a green spell, you draw a card."
//
// Triggers on the cast of each green spell an opponent casts (a multicolour
// spell that is green too). The colour is a word of this card's text (CR 612.2)
// and is read off the spell as it sits on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c77afbeb-cc9b-4770-9a48-e7460562895f",
		Name:         "Insight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.OfColor("G")),
				"Insight — you draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

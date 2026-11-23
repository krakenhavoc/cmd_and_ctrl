package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Didact Echo — Creature — Spirit Cleric {4}{U}, 3/2 (The Lost Caverns of
// Ixalan, #4470):
//
//	"When this creature enters, draw a card.
//	 Descend 4 — This creature has flying as long as there are four or
//	 more permanent cards in your graveyard."
//
// The enters trigger uses the stack and draws one. Descend 4 is an
// ability word (CR 207.2c); the flying is a layer 6 grant (CR 613.1f) on
// the Echo alone, on while its controller's graveyard holds four or more
// permanent cards (CR 110.4) and off the moment it holds fewer
// (SelfKeywordWhileDescend).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8151668c-6459-41b5-9e57-bb47c3876b81",
		Name:         "Didact Echo",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Didact Echo — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Static: []game.StaticAbility{effects.SelfKeywordWhileDescend(4, "flying")},
	})
}

package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nath's Elite — Creature — Elf Warrior {4}{G}, 4/2 (Lorwyn, #5593):
//
//	"All creatures able to block this creature do so.
//	 When this creature enters, clash with an opponent. If you win, put a
//	 +1/+1 counter on this creature. (Each clashing player reveals the top
//	 card of their library, then puts that card on their choice of the top
//	 or bottom. A player wins if their card had a greater mana value.)"
//
// The lure is Breaker of Armies' static (AllAbleToBlockDoSo); the clash
// trigger is Oaken Brawler's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2445fb97-a9c9-4078-8a8e-64b8c3951c4a",
		Name:         "Nath's Elite",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
		Triggered:    []game.TriggeredAbility{effects.OldB543WhenThisEntersClashForCounter("Nath's Elite")},
	})
}

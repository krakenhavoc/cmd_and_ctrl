package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Queen of Dale — Legendary Creature — Human Noble {1}{W}, 2/1:
//
//	"Whenever an opponent casts their first noncreature spell each turn,
//	 you recruit. (Draw a card, then discard a card. If you discarded a
//	 nonland card, create a 1/1 white Human Soldier creature token.)"
//
// Esper Sentinel's trigger condition (each opponent's first noncreature
// spell of the turn, counted by the per-player cast tally), and the
// Queen's controller recruits, on whoever's turn it is.
//
// No simplification. The Hobbit (#4475).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2ec1dd0-86c7-423d-b562-ed95b79bf8f7",
		Name:         "The Queen of Dale",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCastTheirFirstNoncreatureSpellThisTurn,
				"The Queen of Dale — you recruit", effects.Do(effects.Recruit{})),
		},
	})
}

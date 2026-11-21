package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monk of the Open Hand — Creature — Elf Monk {W}, 1/1 (Adventures in the
// Forgotten Realms):
//
//	"Flurry of Blows — Whenever you cast your second spell each turn, put a
//	 +1/+1 counter on this creature."
//
// Flurry of Blows is a flavor word (no rules meaning). The trigger is
// YouCastYourSecondSpellEachTurn, as Devoted Duelist reads it: the cast path
// bumps the per-turn tally before it emits EventCast, so a total of two is
// the second spell, and a spell cast before the Monk entered still counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d3998fc-a68a-4923-a456-b3e43ab6decb",
		Name:         "Monk of the Open Hand",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Monk of the Open Hand — flurry of blows: put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

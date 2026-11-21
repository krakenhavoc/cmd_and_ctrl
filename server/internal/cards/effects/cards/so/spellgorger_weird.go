package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellgorger Weird — Creature — Weird {2}{R}, 2/2 (Ravnica: Clue Edition):
//
//	"Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	 creature."
//
// Academy Instructor's cast trigger over the footprinted CounterOnThis; it fires on casting, so a countered
// spell still grows it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "550811ca-0995-4886-aa8c-5cee118aa54f",
		Name:         "Spellgorger Weird",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Spellgorger Weird — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

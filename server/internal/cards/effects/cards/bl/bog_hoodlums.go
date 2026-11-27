package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Hoodlums — Creature — Goblin Warrior {5}{B}, 4/1 (Lorwyn, #5593):
//
//	"This creature can't block.
//	 When this creature enters, clash with an opponent. If you win, put a
//	 +1/+1 counter on this creature. (Each clashing player reveals the top
//	 card of their library, then puts that card on their choice of the top
//	 or bottom. A player wins if their card had a greater mana value.)"
//
// The restriction is Risen Executioner's RestrictSelf(CantBlock); the clash
// trigger is shared with Paperfin Rascal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c48ce36d-46c3-4432-8711-1b10f12efc90",
		Name:         "Bog Hoodlums",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered:    []game.TriggeredAbility{effects.OldB544WhenThisEntersClashForCounter("Bog Hoodlums")},
	})
}

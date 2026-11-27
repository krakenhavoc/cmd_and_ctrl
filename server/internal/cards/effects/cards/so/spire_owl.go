package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spire Owl — Creature — Bird {1}{U}, 1/1 (Urza's Saga, S82 OLD batch 730,
// #5624):
//
//	"Flying
//	 When this creature enters, look at the top four cards of your library,
//	 then put them back in any order."
//
// Sage Owl's trigger: the four cards are looked at and put back in the order
// the controller chooses; a library of fewer than four shows what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7fa1eca4-0db1-446b-b3cb-15a619ffa97b",
		Name:            "Spire Owl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spire Owl — look at the top four cards of your library, then put them back in any order", effects.Do(effects.LookAtTop{N: 4})),
		},
	})
}

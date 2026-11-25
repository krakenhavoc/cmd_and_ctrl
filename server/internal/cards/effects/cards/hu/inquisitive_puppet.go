package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inquisitive Puppet — Artifact Creature — Construct {1}, 0/2 (Throne of
// Eldraine, S82 old sets):
//
//	"When this creature enters, scry 1.
//	 Exile this creature: Create a 1/1 white Human creature token."
//
// The enters trigger is the ordinary scry. The ability's cost is exiling the
// Puppet itself (Hanged Executioner's), so it has gone before the ability
// resolves and the token is made whatever happens to it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "015004d7-989b-4549-9ffa-b062126d04d2",
		Name:         "Inquisitive Puppet",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inquisitive Puppet — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile this creature: Create a 1/1 white Human creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ExileThis(),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1}),
		}},
	})
}

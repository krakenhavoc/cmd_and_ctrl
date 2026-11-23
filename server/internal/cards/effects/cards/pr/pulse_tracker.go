package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pulse Tracker — Creature — Vampire Rogue {B}, 1/1 (Foundations,
// #4465):
//
//	"Whenever this creature attacks, each opponent loses 1 life."
//
// Postmortem Professor's attack trigger without the life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f916f689-a314-4952-8107-41c4c2f2e811",
		Name:         "Pulse Tracker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Pulse Tracker — each opponent loses 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}

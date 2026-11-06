package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pulse Tracker — Creature — Vampire Rogue {B}, 1/1 (Foundations,
// #4465):
//
//	"Whenever this creature attacks, each opponent loses 1 life."
//
// Postmortem Professor's attack trigger without the life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f916f689-a314-4952-8107-41c4c2f2e811",
		Name:         "Pulse Tracker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Pulse Tracker — each opponent loses 1 life",
				Do(EachOpponentLosesLife{N: 1})),
		},
	})
}

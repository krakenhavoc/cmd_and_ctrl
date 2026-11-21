package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Semester Foreseer // Peer Review — Creature — Human Wizard {3}{U},
// 3/4 // Sorcery {2}{W/U} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 When this creature enters, surveil 1."
//
//	Peer Review — "Create a 2/2 colorless Wizard Soldier creature token
//	 named Cadet. Surveil 1."
//
// The permanent enters already prepared, so its surveil trigger and a
// Peer Review cast in response both find it that way.
//
// No simplification.
func init() {
	const id = "21b8d59f-a90d-44ff-a2e3-9f439ac3e14a"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Semester Foreseer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Semester Foreseer — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Peer Review",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.PeerReviewResolve,
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     id,
		Name:         "Semester Foreseer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Semester Foreseer — surveil 1", Do(Surveil{N: 1})),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Peer Review",
		Completeness: CompletenessFull,
		OnResolve:    peerReviewResolve,
	})
}

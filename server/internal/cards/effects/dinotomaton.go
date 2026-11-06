package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dinotomaton — Artifact Creature — Dinosaur Gnome {3}{R}, 4/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Menace
//	 When this creature enters, target creature you control gains menace
//	 until end of turn."
//
// Menace is the printed keyword (CR 702.111b). The enters trigger targets
// a creature its controller controls as it goes on the stack (CR 603.3d),
// the Dinotomaton itself included, and a target gone by resolution gains
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Dinotomaton — target creature you control gains menace until end of turn"
	Register(Spec{
		OracleID:        "443f2272-6475-4966-8ca2-717d4cf208ee",
		Name:            "Dinotomaton",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, jmpB7FirstLegalTargetGains(label, "menace")),
				targetCreatureYouControl()),
		},
	})
}

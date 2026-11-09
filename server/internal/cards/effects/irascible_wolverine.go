package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Irascible Wolverine — Creature — Wolverine {2}{R}, 3/2 (Outlaws of Thunder Junction):
//
//	"When this creature enters, exile the top card of your library. Until
//	 end of turn, you may play that card.
//	 Plot {2}{R}"
//
// The enters trigger is Outpost Siege's impulse exile (a land included);
// plot is the ordinary special action (CR 702.170), and a plotted cast
// still enters, so the trigger fires then too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "445cd4bd-ba7c-42a2-a794-42aed823d0d3",
		Name:           "Irascible Wolverine",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Irascible Wolverine — exile the top card of your library; until end of turn, you may play it",
				outpostSiegeImpulse),
		},
	})
}

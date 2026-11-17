package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obscura Interceptor — Creature — Octopus Wizard {1}{W}{U}{B}, 3/1
// (Streets of New Capenna, #5491):
//
//	"Flash
//	 Lifelink
//	 When this creature enters, it connives. When it connives this way,
//	 return up to one target spell to its owner's hand. (To have a creature
//	 connive, draw a card, then discard a card. If you discarded a nonland
//	 card, put a +1/+1 counter on that creature.)"
//
// One printed enters trigger whose connive is followed by a reflexive
// trigger (CR 603.12, conniveThenReflexive): the spell target is chosen as
// that trigger goes on the stack, after the draw and discard, and it may be
// no target at all. Returning the spell is not a counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b0870f91-c544-446a-8f7b-a8d03fdb0a2a",
		Name:            "Obscura Interceptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Obscura Interceptor — it connives", obscuraInterceptorConnive),
		},
	})
}

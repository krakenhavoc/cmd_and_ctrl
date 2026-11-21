package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tidespout Tyrant — Creature — Djinn {5}{U}{U}{U}, 5/5 (Ravnica Remastered,
// #5948):
//
//	"Flying
//	 Whenever you cast a spell, return target permanent to its owner's hand."
//
// Drafna's bounce under a cast trigger for any spell (YouCast with no
// predicate). The permanent is chosen as the trigger goes on the stack, so with
// none to target it is removed (CR 603.3d); it may be the Tyrant itself or
// any other permanent, whoever controls it. The trigger resolves before the
// spell that caused it, and returns the permanent only if it is still on the
// battlefield (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "331b44e4-cf08-44f4-8afd-a7374e5dc5ed",
		Name:            "Tidespout Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(nil, "Tidespout Tyrant — return target permanent to its owner's hand", s82Bro19BounceTarget),
				TargetPermanent("target permanent")),
		},
	})
}

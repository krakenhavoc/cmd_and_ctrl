package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "331b44e4-cf08-44f4-8afd-a7374e5dc5ed",
		Name:            "Tidespout Tyrant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(nil, "Tidespout Tyrant — return target permanent to its owner's hand", effects.S82Bro19BounceTarget),
				effects.TargetPermanent("target permanent")),
		},
	})
}

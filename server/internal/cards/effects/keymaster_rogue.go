package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keymaster Rogue — Creature — Human Rogue {3}{U}, 3/2 (Ravnica Remastered,
// #5948):
//
//	"This creature can't be blocked.
//	 When this creature enters, return a creature you control to its owner's
//	 hand."
//
// Cephalid Facetaker's "can't be blocked" static and Whitemane Lion's enters
// trigger: the creature is chosen as the ability resolves, not targeted, and
// may be the Rogue itself, which is the only choice when it is alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8c0c102-f0ef-4ab6-9b1a-ee82ee159c54",
		Name:         "Keymaster Rogue",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Keymaster Rogue — return a creature you control", Do(ReturnOneYouControl{
				Match:    MatchCreature,
				Question: "Keymaster Rogue — return a creature you control to its owner's hand",
			})),
		},
	})
}

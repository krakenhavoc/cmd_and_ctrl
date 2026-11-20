package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riftwing Cloudskate — Creature — Illusion {3}{U}{U}, 2/2 (Time Spiral, #5596):
//
//	"Flying
//	 When this creature enters, return target permanent to its owner's
//	 hand.
//	 Suspend 3—{1}{U} (Rather than cast this card from your hand, you
//	 may pay {1}{U} and exile it with three time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// Man-o'-War's mandatory enters trigger aimed at any permanent, on a flier
// that is usually suspended for {1}{U} on turn two. Suspend (CR 702.62a)
// is the engine's: the countdown, the free cast and the haste are grown
// from the declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae2e6ffd-fb93-4c47-88e7-e84460eefbe5",
		Name:            "Riftwing Cloudskate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions: []game.SpecialAction{
			Suspend(3, "{1}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Riftwing Cloudskate — return target permanent to its owner's hand",
				bounceTheTargetToHand), TargetPermanent("target permanent")),
		},
	})
}

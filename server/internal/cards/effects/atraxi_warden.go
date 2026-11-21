package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atraxi Warden — Creature — Alien Eye {4}{W}{W}, 6/6 (Doctor Who, #5948):
//
//	"Flying
//	 When this creature enters, exile up to one target tapped creature.
//	 Suspend 5—{1}{W} (Rather than cast this card from your hand, you may pay
//	 {1}{W} and exile it with five time counters on it. At the beginning of
//	 your upkeep, remove a time counter. When the last is removed, you may cast
//	 it without paying its mana cost. It has haste.)"
//
// Flying is a printed keyword and suspend is the keyword's special action
// (Suspend). The entry trigger has an optional target, a tapped creature of
// any controller, judged on announcement and again as it resolves (CR 608.2b),
// and exiles it (permanently: no return).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5f31a2d6-649a-40f4-89ce-95c635340e9e",
		Name:            "Atraxi Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Suspend(5, "{1}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Atraxi Warden — exile up to one target tapped creature", gapB91ExileTheChosenTarget),
				UpToOneTargetCreature("up to one target tapped creature", tappedPermanent())),
		},
	})
}

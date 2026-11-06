package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prayer of Binding — Enchantment {3}{W} (Foundations, #4465):
//
//	"Flash
//	 When this enchantment enters, exile up to one target nonland
//	 permanent an opponent controls until this enchantment leaves the
//	 battlefield. You gain 2 life."
//
// Makeshift Binding with Earth Kingdom Jailer's "up to one" over any
// nonland permanent. The exile is exileChosenTargetUntilThisLeaves: the
// permanent returns under its owner's control when the Prayer leaves, and
// nothing is exiled if the Prayer left before the trigger resolved. The
// 2 life follows either way, and with no target chosen it is all the
// trigger does. A chosen target that is gone by resolution leaves the
// trigger with no legal target, so it does nothing, the life included.
// Flash rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4505ff9f-97f1-45d1-aea9-5590fefb4a4d",
		Name:            "Prayer of Binding",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Prayer of Binding — exile up to one target nonland permanent an opponent controls until this enchantment leaves the battlefield; you gain 2 life",
				exileChosenTargetUntilThisLeavesThenGainLife(
					"Prayer of Binding — the exiled permanent returns when Prayer of Binding leaves the battlefield", 2)),
				TargetPermanent("up to one target nonland permanent an opponent controls", Nonland(), OpponentControls()).WithCount(0, 1)),
		},
	})
}

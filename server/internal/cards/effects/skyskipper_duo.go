package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyskipper Duo — Creature — Bird Frog {4}{U}, 3/3 (Bloomburrow, #4467):
//
//	"Flying
//	 When this creature enters, exile up to one other target creature you
//	 control. Return it to the battlefield under its owner's control at
//	 the beginning of the next end step."
//
// Chief Kyle, Transporter Tech's delayed blink on an entry trigger
// (exileTargetsThenScheduleReturn): the card returns under its owner's
// control as a new object at the next end step (CR 400.7, CR 603.7),
// and a token exiled this way does not return. Choosing no target
// exiles nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a50db36c-948b-4132-a5e5-84aab1f42612",
		Name:            "Skyskipper Duo",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Skyskipper Duo — exile up to one other target creature you control; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return exileTargetsThenScheduleReturn(NewContext(g, item), "Skyskipper Duo — return the exiled creature")
					}),
				Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
			),
		},
	})
}

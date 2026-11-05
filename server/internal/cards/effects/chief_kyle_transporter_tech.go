package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chief Kyle, Transporter Tech — Legendary Creature — Human Artificer
// {2}{W}{U}, 3/2 (Star Trek Commander, #4974):
//
//	"Landfall — Whenever a land you control enters, exile up to one other
//	 target creature you control. Return that card to the battlefield
//	 under its owner's control at the beginning of the next end step."
//
// Landfall is an ability word. Each land that enters under Kyle's
// controller's control triggers once. The body is Ennis's and Roon's
// delayed blink (exileTargetsThenScheduleReturn) over an "up to one other"
// target: with no target nothing is exiled and nothing is scheduled, a
// target that became illegal stays where it is (CR 608.2b), and the card
// returns under its owner's control as a new object at the next end step
// (CR 400.7, CR 603.7); a token exiled this way does not return (CR 111.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02583549-b618-4d41-9f46-209a99e30cd7",
		Name:         "Chief Kyle, Transporter Tech",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Landfall("Chief Kyle, Transporter Tech — exile up to one other target creature you control; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return exileTargetsThenScheduleReturn(NewContext(g, item),
							"Chief Kyle, Transporter Tech — return the exiled creature")
					}),
				Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
			),
		},
	})
}

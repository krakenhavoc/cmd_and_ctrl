package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// K'un-Lun Warrior — Creature — Human Warrior Hero {1}{R}, 2/2:
//
//	"When this creature enters, you may sacrifice an artifact or discard
//	 a card. If you do, draw a card."
//
// #4566's either/or may-pay (SacrificeAnArtifactOrDiscard): one prompt,
// the player names the branch, a branch they can't pay is not offered,
// and the draw happens only once the chosen branch is paid (CR 118.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0618a217-2a6e-4fa1-bfe6-89c6612a594d",
		Name:         "K'un-Lun Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("K'un-Lun Warrior — you may sacrifice an artifact or discard a card; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return mayPayArtifactOrCard(NewContext(g, item),
						"K'un-Lun Warrior — sacrifice an artifact or discard a card to draw a card?", mayPayDrawACard)
				}),
		},
	})
}

// mayPayArtifactOrCard is "you may sacrifice an artifact or discard a
// card. If you do, <then>" (#4566): K'un-Lun Warrior, Contract Hero and
// Vision of Love.
func mayPayArtifactOrCard(ctx *Context, question string, then game.PayUnlessElse, objects ...game.ObjectRef) error {
	return MayPay{
		Chooser:      ctx.Controller(),
		Pay:          SacrificeAnArtifactOrDiscard(false),
		Question:     question,
		Then:         then,
		CarryObjects: objects,
	}.Apply(ctx)
}

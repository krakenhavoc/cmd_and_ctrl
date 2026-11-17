package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gitaxian Anatomist — Creature — Phyrexian Wizard {3}{U}, 2/5 (Phyrexia: All
// Will Be One, S82):
//
//	"When this creature enters, you may tap it. If you do, proliferate."
//
// The may-tap is asked as the trigger resolves; "if you do" holds only when
// the Anatomist is still on the battlefield and untapped, so one that has
// left, or that entered tapped, does not proliferate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa9bde06-098c-41b3-b816-49daacc02c50",
		Name:         "Gitaxian Anatomist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gitaxian Anatomist — you may tap it; if you do, proliferate",
				Do(MayChoice{
					Question: "Gitaxian Anatomist — tap it to proliferate?",
					YesLabel: "Tap",
					NoLabel:  "Decline",
					Then:     gitaxianAnatomistAnswer,
				})),
		},
	})
}

// gitaxianAnatomistAnswer is the "yes": tap the Anatomist, and proliferate if
// the tap happened. Keyed.
var gitaxianAnatomistAnswer = MayYes("confirm/gitaxian-anatomist-tap", func(ctx *Context, _ game.ConfirmAnswer) error {
	info, ok := ctx.SourcePermanent()
	if !ok || info.Left || info.Tapped || ctx.isNewSourceObject(ctx.Source()) {
		return nil
	}
	if err := (TapTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
		return err
	}
	return Proliferate{}.Apply(ctx)
})

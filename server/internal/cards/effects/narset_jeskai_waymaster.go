package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narset, Jeskai Waymaster — Legendary Creature — Human Monk {U}{R}{W},
// 3/4 (Tarkir: Dragonstorm, #4463):
//
//	"At the beginning of your end step, you may discard your hand. If
//	 you do, draw cards equal to the number of spells you've cast this
//	 turn."
//
// Smellerbee's keyed yes/no at the end step, with one difference the
// printed ruling makes: you may discard your hand even if it has no
// cards in it, and you still draw. So a yes always counts as "if you
// do", and an empty hand is not a reason to skip the draw. The count is
// the turn's cast tally (CastTally.Total), read as the trigger resolves
// (CR 608.2h): every spell its controller cast this turn, whatever has
// happened to it since; a copy is not cast and does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a2a74b8-0689-4959-9348-02cc832194cf",
		Name:         "Narset, Jeskai Waymaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Narset, Jeskai Waymaster — you may discard your hand; if you do, draw a card for each spell you've cast this turn",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Narset, Jeskai Waymaster — discard your hand? If you do, draw cards equal to the number of spells you've cast this turn.",
						YesLabel: "Discard and draw",
						NoLabel:  "Keep your hand",
						Then:     narsetJeskaiWaymasterDiscardHand,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// narsetJeskaiWaymasterDiscardHand is the yes. An on-disk key: never
// renamed, never reused.
var narsetJeskaiWaymasterDiscardHand = MayYes("confirm/narset-jeskai-waymaster-discard-hand", func(ctx *Context, r game.ConfirmAnswer) error {
	if _, err := discardWholeHand(ctx.Game, r.Chooser); err != nil {
		return err
	}
	return DrawCards{Player: r.Chooser, N: ctx.Game.CastTallyFor(r.Chooser).Total}.Apply(ctx)
})

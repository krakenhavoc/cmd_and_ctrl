package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oblivious Bookworm — Creature — Human Wizard {G}{U}, 2/3 (Duskmourn:
// House of Horror, #5241):
//
//	"At the beginning of your end step, you may draw a card. If you do,
//	 discard a card unless a permanent entered the battlefield face down
//	 under your control this turn or you turned a permanent face up this
//	 turn."
//
// The "unless" is read as the ability resolves (the ruling: by then it
// is too late to put a permanent face down or turn one up), off two
// turn-tally facts: a permanent that entered face down under your
// control (a face-down spell resolving, a manifest, a cloak; one turned
// face up since still counts) and a permanent you turned face up, by
// the special action or by an effect of yours. The yes is keyed (#4291);
// declining the draw declines the discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "300fd0a2-23e6-4da9-bfaa-ad626097d5b5",
		Name:         "Oblivious Bookworm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Oblivious Bookworm — you may draw a card; if you do, discard a card unless you put a permanent face down or turned one face up this turn",
				func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Oblivious Bookworm — draw a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     obliviousBookwormDraw,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// obliviousBookwormDraw is the yes: draw, then discard unless the turn
// saw a face-down entry or a face-up turn of yours.
var obliviousBookwormDraw = MayYes("confirm/oblivious-bookworm-draw", func(ctx *Context, _ game.ConfirmAnswer) error {
	you := ctx.Controller()
	if obliviousBookwormSpared(ctx.Game, you) {
		return DrawCards{Player: you, N: 1}.Apply(ctx)
	}
	return drawThenDiscard(ctx.Game, you, ctx.Source(), 1, 1, "")
})

// obliviousBookwormSpared is the "unless": a permanent entered the
// battlefield face down under your control this turn, or you turned a
// permanent face up this turn.
func obliviousBookwormSpared(g *game.Game, you uuid.UUID) bool {
	return g.PermanentEnteredFaceDownThisTurn(you) || g.TurnedPermanentFaceUpThisTurn(you)
}

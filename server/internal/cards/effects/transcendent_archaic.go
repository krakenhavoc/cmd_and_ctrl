package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transcendent Archaic — Creature — Avatar {7}, 6/6 (Secrets of Strixhaven,
// #4456):
//
//	"Vigilance
//	 Converge — When this creature enters, you may draw X cards, where X is
//	 the number of colors of mana spent to cast this spell. If you draw one
//	 or more cards this way, discard two cards."
//
// Converge is an ability word (CR 207.2c). X is the number of distinct colours
// among the mana that paid for the spell this permanent was, read as the
// trigger is put on the stack (the permanent's record of its cast, so a
// payment of {W}{U}{B} plus generic is three). WantsDistinctColors pays the
// generic cost with colours not yet spent when it can, and ReadsOwnPayment
// makes a waived payment ask what was paid. A creature that was not cast has
// spent nothing, so X is zero and nothing is asked.
//
// "You may" is asked as the trigger resolves, and the draw is declared to the
// engine, so a player a draw limit stops is not offered it. The discard is the
// draw's continuation (drawThenDiscard): it opens only after every draw has
// resolved (a draw paused on a replacement prompt holds it), and the drawn
// cards are legal discards. An empty library draws nothing, so nothing is
// discarded.
//
// No simplification.
func init() {
	const label = "Transcendent Archaic — Converge: you may draw X cards (X is the number of colors of mana spent); if you do, discard two cards"
	enters := WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		x := item.Params.Amount
		if x <= 0 {
			return nil
		}
		return MayChoice{
			Question: "Transcendent Archaic — draw " + strconv.Itoa(x) + " cards, then discard two cards?",
			YesLabel: "Draw " + strconv.Itoa(x),
			NoLabel:  "Decline",
			Draws:    x,
			Word:     strconv.Itoa(x),
			Then:     socB38ArchaicDraw,
		}.Apply(NewContext(g, item))
	})
	enters.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = source.ManaSpentToCast().ColorCount()
		return item
	}
	Register(Spec{
		OracleID:            "fdd4b3a9-83ce-41bf-82e2-7808657e2c09",
		Name:                "Transcendent Archaic",
		Completeness:        CompletenessFull,
		PrintedKeywords:     []string{"vigilance"},
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Triggered:           []game.TriggeredAbility{enters},
	})
}

// socB38ArchaicDraw is the "yes" of Transcendent Archaic's trigger: draw X
// (the Word the prompt carries), then discard two once the draws are done.
var socB38ArchaicDraw = MayYes("confirm/transcendent-archaic-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	return drawThenDiscard(ctx.Game, r.Chooser, ctx.Source(), wordInt(r), 2, "")
})

package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forgotten Creation — Creature — Zombie Horror {3}{U}, 3/3:
//
//	"Skulk (This creature can't be blocked by creatures with greater
//	 power.)
//	 At the beginning of your upkeep, you may discard all the cards in
//	 your hand. If you do, draw that many cards."
//
// Skulk is an enforced evasion keyword (block_legality.go), so the card
// file only names it.
//
// The upkeep ability is a "you may" on the TRIGGER (Optional), asked as
// it resolves; a yes discards the whole hand and the draw follows. "If
// you do" is same-resolution sequencing, not a CR 603.12 reflexive
// trigger (see Jackdaw), so nothing is put on the stack between the two
// halves. "That many" is the number of cards that REALLY reached a
// graveyard, read off the discard's own continuation
// (DiscardRandomThenForEffect hands it the cards discarded), not the
// size of the hand beforehand: a discard that was blocked, or a card a
// replacement kept in hand, is not counted, and a discard that pauses
// for a prompt (madness, a commander's CR 903.9a offer) draws only once
// it has settled. "At random" in that primitive's name is the order, not
// the choice: the whole hand goes, so there is nothing for the player
// to pick.
//
// An empty hand discards nothing and draws nothing, which is a legal
// thing to say yes to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c26964c7-4259-4281-b450-f0c2ed34b37c",
		Name:            "Forgotten Creation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			Optional(
				AtYourUpkeep("Forgotten Creation — discard your hand, then draw that many cards", forgottenCreationWheel),
				"Forgotten Creation — discard your hand? (Draw that many cards.)"),
		},
	})
}

// forgottenCreationWheel discards the controller's whole hand and then
// draws as many cards as were discarded.
func forgottenCreationWheel(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	return g.DiscardRandomKeyedForEffect(item.Controller, p.Hand.Size(), NewContext(g, item).ItemTail(forgottenCreationDraw))
}

// forgottenCreationDraw draws a card per card discarded (#4253 slice 9).
var forgottenCreationDraw = CardTail("forgotten-creation/draw-per-discarded", func(ctx *Context, o game.TailOutcome) error {
	g, item, discarded := ctx.Game, ctx.Item, o.IDs

	return DrawCards{Player: item.Controller, N: len(discarded)}.Apply(NewContext(g, item))
})

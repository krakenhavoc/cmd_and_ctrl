package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB081FactOrFictionReveal is Fact or Fiction's sentence on a creature's enters
// trigger (Sphinx of Uthuun, #5507): reveal the top n cards of your library,
// an opponent separates them into two piles, you put one pile into your hand
// and the other into your graveyard. Fact or Fiction's own helpers do the rest: the opponent who
// separates is the controller's choice among several, and a card that has
// moved since is skipped.
func OldB081FactOrFictionReveal(ctx *Context, n int) error {
	const card = "Sphinx of Uthuun"
	revealed := ctx.Game.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), n,
		card+" — reveal the top cards of your library")
	if len(revealed) == 0 {
		return nil
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: card + " — choose an opponent to separate these cards into two piles",
		ThenKey:  oldB081Split,
		Carry:    revealed,
	}.Apply(ctx)
}

// oldB081Split hands the revealed cards to the chosen opponent. An on-disk
// key: never renamed, never reused.
var oldB081Split = ChoosePlayerThen("option-pick/old-b081-fact-or-fiction-splitter", func(ctx *Context, splitter uuid.UUID, revealed []uuid.UUID) error {
	if splitter == uuid.Nil {
		// Nobody left to separate them (CR 800.4a): one pile, taken.
		return FactOrFictionSettle(ctx.Game, revealed, nil)
	}
	return PileSplit{
		Splitter:      splitter,
		Chooser:       ctx.Controller(),
		Owner:         ctx.Controller(),
		SplitQuestion: "Sphinx of Uthuun — separate these cards into two piles",
		PickQuestion:  "Sphinx of Uthuun — take one pile into your hand; the other goes to your graveyard",
		Cards:         revealed,
		ThenTail:      ctx.Tail(oldB081Take),
	}.Apply(ctx)
})

// oldB081Take puts the chosen pile into the hand and the other into the
// graveyard. An on-disk key.
var oldB081Take = PileTail("then/old-b081-fact-or-fiction-take", func(ctx *Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
	return FactOrFictionSettle(ctx.Game, taken, left)
})

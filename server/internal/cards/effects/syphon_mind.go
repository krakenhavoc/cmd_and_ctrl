package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syphon Mind — Sorcery {3}{B} (EDHREC rank 1049):
//
//	"Each other player discards a card. You draw a card for each card
//	 discarded this way."
//
// The multiplayer Sign in Blood: three cards for four mana at a full
// table. Each opponent's discard is their own choice, one prompt per
// opponent over their own hand, so nobody picks for anyone else.
//
// The catalog's one card that reads discard-FIRST-then-draw, which is
// why the draw rides DiscardPrompt.Then (#651): "a card for each card
// discarded THIS WAY" counts what actually reached a graveyard, so
// each opponent's own answer is what buys the caster a card. Before
// the discard was a real prompt the count had to be settled up front
// from who was holding cards, and the draw happened while every
// opponent was still deciding; that was this card's declared caveat
// and it is gone. An opponent who leaves the game before answering
// now buys nothing, which is the printed card.
//
// An opponent with an empty hand is skipped rather than prompted —
// they discard nothing, so there is nothing for the caster to draw
// off. (The guard belongs here and not in the engine: the empty-hand
// case of a discard prompt still runs Then, because "discard your
// hand, THEN draw three" draws three from an empty hand. This card's
// "then" is a payoff counting the discard, not the next instruction.)
//
// Sandbox simplification, cosmetic: the caster's cards arrive one per
// answer rather than all at once once the table has discarded. No
// priority passes in between and nothing in the catalog can see the
// gap — a draw payoff counts the same draws either way.
func init() {
	Register(Spec{
		OracleID:     "abc37d6c-6300-47b5-a679-9db5b83eb54f",
		Name:         "Syphon Mind",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			caster := item.Controller
			source := item.SourceCardID
			for _, opp := range ctx.Opponents() {
				if p := ctx.PlayerByID(opp); p == nil || p.Hand.Size() == 0 {
					continue
				}
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player: opp,
					Source: source,
					N:      1,
					Then: func(g *game.Game) error {
						return g.DrawNForEffect(caster, 1)
					},
				})
			}
			return nil
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vicious Rumors — Sorcery {B} (EDHREC rank 4389):
//
//	"Vicious Rumors deals 1 damage to each opponent. Each opponent
//	 discards a card, then mills a card. You gain 1 life."
//
// Four small things for one black mana, and in a four-player game
// every one of the first three happens three times. It is a Commander
// card by accident of arithmetic: nine points of value off a card
// that was printed as filler. Roadmap batch 42 (#449), "no new
// machinery".
//
// The discard is a real CR 701.8a choice — each opponent picks, in
// their own prompt — rather than the "at random" primitive. That
// distinction is #651's rule and it is load-bearing here: an opponent
// choosing which card to pitch is a meaningfully different card from
// one losing a random one.
//
// Order is printed and kept: damage, then discard, then mill, then
// your life gain. The mill rides the discard prompt's Then rather
// than sitting on the next line, because the prompt is queued and
// answered later — a mill written as the next statement would happen
// BEFORE the opponent chose, and the card they milled would be a card
// they could no longer discard. Then also fires immediately for an
// opponent with an empty hand, which is right: CR 701.8a discards as
// many as you can (none), and "then mills a card" is not conditional
// on there having been a card to pitch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55b72b1f-6463-406a-8824-d99a3c028285",
		Name:         "Vicious Rumors",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageToEachOpponent(ctx.Game, item, 1); err != nil {
				return err
			}
			for _, opp := range ctx.Opponents() {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   opp,
					Source:   ctx.Source(),
					N:        1,
					Question: "Vicious Rumors — discard a card, then mill a card",
					Then: func(g *game.Game) error {
						return g.MillNForEffect(opp, 1)
					},
				})
			}
			return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
		},
	})
}

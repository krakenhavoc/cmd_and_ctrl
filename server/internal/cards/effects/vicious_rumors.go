package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
// your life gain — and the last of those four took #1027 to get right.
//
// TWO CONTINUATIONS, and the difference between them is the whole
// shape (prompt_run.go). The mill is PER OPPONENT and rides that
// opponent's own prompt: "…discards a card, THEN mills a card" is one
// sentence about one player, so a mill written as the next statement
// would happen BEFORE they chose, and the card they milled would be a
// card they could no longer discard. DiscardPrompt.Then also fires
// immediately for an opponent with an empty hand, which is right:
// CR 701.8a discards as many as you can (none), and "then mills a
// card" is not conditional on there having been a card to pitch.
//
// The LIFE GAIN is a separate printed sentence about the caster, so it
// belongs to the whole instruction: it is the RUN's continuation, and
// it runs once, after every opponent has answered. It used to sit on
// the line after the fan-out, which put the caster a life ahead while
// the table was still deciding — a clause about a PLAYER that happens
// either way (ADR 0013 §5m item 5), so the run's answer is
// deliberately ignored and what the continuation buys is the order.
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
			caster := ctx.Controller()
			return ctx.Game.EachPlayerDiscardsThenForEffect(caster,
				game.DiscardPrompt{
					Source:   ctx.Source(),
					N:        1,
					Question: "Vicious Rumors — discard a card, then mill a card",
					Then:     viciousRumorsMillTheSeatJustAsked,
				},
				func(g *game.Game, _ game.PromptedDiscards) error {
					return GainLife{Player: caster, Amount: 1}.Apply(NewContext(g, item))
				})
		},
	})
}

// viciousRumorsMillTheSeatJustAsked is the per-opponent "then mills a
// card".
//
// The seat is the leg's own argument rather than something captured,
// which is what makes the fan-out safe: the prompt template is ONE
// value copied per seat (discardRunLocked), so a closure over a loop
// variable would mill whichever player the loop stopped on. What the
// leg is handed is the seat whose discard has just finished, which is
// the player this sentence is about.
//
// Caller holds g.mu.
func viciousRumorsMillTheSeatJustAsked(g *game.Game, seat uuid.UUID, _ []uuid.UUID) error {
	return g.MillNForEffect(seat, 1)
}
